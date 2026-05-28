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

	"rag/ragclient"

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
		AbstractText  string `json:"AbstractText"`
		AbstractURL   string `json:"AbstractURL"`
		Heading       string `json:"Heading"`
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

// KnowledgeSearchTool searches the user's knowledge bases via RAG service
type KnowledgeSearchTool struct {
	ragClient ragclient.Rag
	kbIDs     []string // Authorized KB IDs for scoped search (optional)
}

func NewKnowledgeSearchTool(ragClient ragclient.Rag, kbIDs []string) *KnowledgeSearchTool {
	return &KnowledgeSearchTool{ragClient: ragClient, kbIDs: kbIDs}
}

func (t *KnowledgeSearchTool) Name() string {
	return "knowledge_search"
}

func (t *KnowledgeSearchTool) Description() string {
	return "搜索用户知识库中的文档内容。参数: query (搜索问题), kb_id (可选, 指定知识库ID, 不指定则搜索所有授权的知识库), top_k (可选, 返回结果数, 默认3)"
}

func (t *KnowledgeSearchTool) Execute(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("缺少搜索问题")
	}

	if t.ragClient == nil {
		return "知识库服务不可用，请检查 RAG 服务配置。", nil
	}

	topK := int32(3)
	if tk, ok := args["top_k"].(float64); ok {
		topK = int32(tk)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kbID := ""
	if id, ok := args["kb_id"].(string); ok {
		kbID = id
	}

	if kbID == "" {
		// Search across authorized knowledge bases
		resp, err := t.ragClient.CrossKBSearch(ctx, &ragclient.CrossKBSearchReq{
			Query: query,
			TopK:  topK,
			KbIds: t.kbIDs,
		})
		if err != nil {
			logx.Errorf("cross KB search failed: %v", err)
			return "", fmt.Errorf("知识库搜索失败: %w", err)
		}
		return t.formatSearchResults(resp.Items, query), nil
	}

	resp, err := t.ragClient.SearchKnowledge(ctx, &ragclient.SearchKnowledgeReq{
		KbId:  kbID,
		Query: query,
		TopK:  topK * 2, // Request more for better fusion
	})
	if err != nil {
		logx.Errorf("knowledge search failed: %v", err)
		return "", fmt.Errorf("知识库搜索失败: %w", err)
	}
	return t.formatSearchResults(resp.Items, query), nil
}

func (t *KnowledgeSearchTool) formatSearchResults(results []*ragclient.RecallItem, query string) string {
	if len(results) == 0 {
		return fmt.Sprintf("未在知识库中找到与 \"%s\" 相关的内容。", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("找到 %d 条相关知识:\n\n", len(results)))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("### [%d] %s (相关度: %.0f%%)\n", i+1, r.DocName, r.FusionScore*100))
		sb.WriteString(fmt.Sprintf("%s\n", r.Content))
		if len(r.Entities) > 0 {
			sb.WriteString(fmt.Sprintf("\n相关实体: %s\n", strings.Join(r.Entities, ", ")))
		}
		sb.WriteString("\n")
	}
	return sb.String()
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

// SkillDefinition represents a tool/skill from a bot template's tool_definitions JSON.
type SkillDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// ToolExecutor runs tools requested by LLM
type ToolExecutor struct {
	registry      *ToolRegistry
	ragClient     ragclient.Rag
	kbIDs         []string
	customSkillDefs []SkillDefinition
}

func NewToolExecutor(ragClient ragclient.Rag, kbIDs []string) *ToolExecutor {
	reg := NewToolRegistry()
	reg.Register(NewWebSearchTool())
	if ragClient != nil {
		reg.Register(NewKnowledgeSearchTool(ragClient, kbIDs))
	}
	return &ToolExecutor{
		registry:  reg,
		ragClient: ragClient,
		kbIDs:     kbIDs,
	}
}

// LoadCustomTools parses tool_definitions JSON from a bot template and registers custom skills.
func (e *ToolExecutor) LoadCustomTools(defsJSON string) {
	if defsJSON == "" {
		return
	}

	var defs []SkillDefinition
	if err := json.Unmarshal([]byte(defsJSON), &defs); err != nil {
		logx.Errorf("parse tool_definitions JSON failed: %v", err)
		return
	}

	e.customSkillDefs = defs
	for _, def := range defs {
		if def.Name == "" {
			continue
		}
		if e.registry.Get(def.Name) != nil {
			continue
		}
		var tool Tool
		switch def.Type {
		case "knowledge_search":
			if e.ragClient != nil {
				kbIDs := e.kbIDs
				if rawIDs, ok := def.Config["kb_ids"]; ok {
					if ids, ok := rawIDs.([]interface{}); ok {
						kbIDs = nil
						for _, id := range ids {
							if s, ok := id.(string); ok {
								kbIDs = append(kbIDs, s)
							}
						}
					}
				}
				tool = NewKnowledgeSearchTool(e.ragClient, kbIDs)
			}
		case "web_search":
			tool = NewWebSearchTool()
		default:
			tool = &configurableSkill{def: def}
		}
		if tool != nil {
			e.registry.Register(tool)
		}
	}
	logx.Infof("tool_executor: loaded %d custom tools from template", len(defs))
}

// GetAvailableToolsList returns all registered tools for the system prompt.
func (e *ToolExecutor) GetAvailableToolsList() []Tool {
	var tools []Tool
	for _, t := range e.registry.List() {
		if tool := e.registry.Get(t["name"]); tool != nil {
			tools = append(tools, tool)
		}
	}
	return tools
}

// configurableSkill is a generic tool wrapper for custom skill definitions.
type configurableSkill struct {
	def SkillDefinition
}

func (s *configurableSkill) Name() string {
	return s.def.Name
}

func (s *configurableSkill) Description() string {
	return s.def.Description
}

func (s *configurableSkill) Execute(args map[string]interface{}) (string, error) {
	argsJSON, _ := json.Marshal(args)
	return fmt.Sprintf("技能 %s 收到参数: %s。此技能需要外部服务支持，当前为占位实现。请根据技能描述给出合理回复。", s.def.Name, string(argsJSON)), nil
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
