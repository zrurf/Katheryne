package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// Tool represents a callable MCP-like tool
type Tool interface {
	Name() string
	Description() string
	Execute(args map[string]interface{}) (string, error)
}

// WebSearchResult represents a single search result
type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearchTool implements web search via DuckDuckGo API
type WebSearchTool struct {
	client *http.Client
}

func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *WebSearchTool) Name() string {
	return "web_search"
}

func (t *WebSearchTool) Description() string {
	return "搜索互联网获取实时信息。参数: query (搜索关键词), max_results (最大结果数, 默认5)"
}

func (t *WebSearchTool) Execute(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("缺少搜索关键词")
	}

	maxResults := 5
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	return t.search(query, maxResults)
}

func (t *WebSearchTool) search(query string, maxResults int) (string, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	resp, err := t.client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("搜索请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取搜索结果失败: %w", err)
	}

	var result struct {
		AbstractText string `json:"AbstractText"`
		AbstractURL  string `json:"AbstractURL"`
		Heading      string `json:"Heading"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		logx.Errorf("parse search result: %v, body: %s", err, truncate(string(body), 200))
		return "", fmt.Errorf("解析搜索结果失败")
	}

	var sb strings.Builder
	if result.AbstractText != "" {
		sb.WriteString(fmt.Sprintf("## %s\n\n", result.Heading))
		sb.WriteString(result.AbstractText)
		sb.WriteString(fmt.Sprintf("\n\n来源: %s\n", result.AbstractURL))
	}

	count := 0
	if result.AbstractText != "" {
		count = 1
	}

	for _, topic := range result.RelatedTopics {
		if count >= maxResults {
			break
		}
		sb.WriteString(fmt.Sprintf("\n- %s", topic.Text))
		if topic.FirstURL != "" {
			sb.WriteString(fmt.Sprintf(" [链接](%s)", topic.FirstURL))
		}
		count++
	}

	if sb.Len() == 0 {
		return fmt.Sprintf("未找到关于 \"%s\" 的搜索结果。", query), nil
	}

	return sb.String(), nil
}

// ToolRegistry manages available tools
type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (r *ToolRegistry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *ToolRegistry) Get(name string) Tool {
	return r.tools[name]
}

func (r *ToolRegistry) List() []map[string]string {
	var tools []map[string]string
	for _, t := range r.tools {
		tools = append(tools, map[string]string{
			"name":        t.Name(),
			"description": t.Description(),
		})
	}
	return tools
}

// ToolExecutor runs tools requested by LLM
type ToolExecutor struct {
	registry *ToolRegistry
}

func NewToolExecutor() *ToolExecutor {
	reg := NewToolRegistry()
	reg.Register(NewWebSearchTool())
	return &ToolExecutor{registry: reg}
}

func (e *ToolExecutor) ExecuteToolCalls(ctx context.Context, response string) (replaced string, hasTools bool) {
	hasTools = false
	replaced = response

	if !strings.Contains(response, "<tool_call>") {
		return response, false
	}

	for {
		start := strings.Index(replaced, "<tool_call>")
		if start < 0 {
			break
		}
		end := strings.Index(replaced, "</tool_call>")
		if end < 0 {
			break
		}

		jsonStr := replaced[start+len("<tool_call>") : end]
		var toolCall struct {
			Name string                 `json:"name"`
			Args map[string]interface{} `json:"args"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &toolCall); err != nil {
			logx.Errorf("parse tool_call JSON: %v, json: %s", err, jsonStr)
			replaced = strings.Replace(replaced, replaced[start:end+len("</tool_call>")],
				fmt.Sprintf("[Tool调用失败: 参数解析错误]"), 1)
			hasTools = true
			continue
		}

		tool := e.registry.Get(toolCall.Name)
		if tool == nil {
			replaced = strings.Replace(replaced, replaced[start:end+len("</tool_call>")],
				fmt.Sprintf("[未知工具: %s]", toolCall.Name), 1)
			hasTools = true
			continue
		}

		result, err := tool.Execute(toolCall.Args)
		if err != nil {
			logx.Errorf("execute tool %s: %v", toolCall.Name, err)
			replaced = strings.Replace(replaced, replaced[start:end+len("</tool_call>")],
				fmt.Sprintf("[Tool执行失败: %s]", err.Error()), 1)
			hasTools = true
			continue
		}

		replaced = strings.Replace(replaced, replaced[start:end+len("</tool_call>")],
			fmt.Sprintf("\n[%s 执行结果]\n%s\n", toolCall.Name, result), 1)
		hasTools = true
	}

	return replaced, hasTools
}

// GetAvailableToolsDescription returns a description of available tools for system prompt
func (e *ToolExecutor) GetAvailableToolsDescription() string {
	tools := e.registry.List()
	if len(tools) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("你可以使用以下工具来获取实时信息:\n")
	for _, t := range tools {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", t["name"], t["description"]))
	}
	sb.WriteString("\n使用工具时，请按以下格式输出:\n")
	sb.WriteString("<tool_call>{\"name\": \"tool_name\", \"args\": {\"key\": \"value\"}}</tool_call>\n")
	return sb.String()
}