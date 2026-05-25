package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"rag/internal/dao"
)

// FeishuSourceConfig represents the source_config JSON for Feishu knowledge bases.
type FeishuSourceConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	WikiToken string `json:"wiki_token"` // Wiki space node token
	WikiType  string `json:"wiki_type"`  // "wiki" or "drive" (drive = 云空间)
	NodeToken string `json:"node_token"` // Specific folder/node to sync from (optional; defaults to wiki_token root)
}

// feishuAccessTokenResp is the response from Feishu token API.
type feishuAccessTokenResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccessToken string `json:"access_token"`
		Expire      int64  `json:"expire"`
	} `json:"tenant_access_token,omitempty"`
}

// feishuNodeInfoResp represents a node info response.
type feishuNodeInfoResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Node struct {
			Token      string `json:"token"`
			ObjType    string `json:"obj_type"` // doc, docx, sheet, bitable, folder, mindnote, etc.
			ParentToken string `json:"parent_token"`
			NodeType   string `json:"node_type"` // origin, shortcut
			Title      string `json:"title"`
			HasChild   bool   `json:"has_child"`
			Owner      string `json:"owner"`
			CreateTime string `json:"create_time"`
			UpdateTime string `json:"update_time"`
			URL        string `json:"url"`
		} `json:"node"`
	} `json:"data"`
}

// feishuChildrenResp represents children listing response.
type feishuChildrenResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Items     []feishuNodeItem `json:"items"`
		PageToken string           `json:"page_token"`
		HasMore   bool             `json:"has_more"`
	} `json:"data"`
}

type feishuNodeItem struct {
	Token      string `json:"token"`
	ObjType    string `json:"obj_type"`
	ParentToken string `json:"parent_token"`
	NodeType   string `json:"node_type"`
	Title      string `json:"title"`
	HasChild   bool   `json:"has_child"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	Owner      string `json:"owner"`
}

// feishuDocRawResp represents raw document content response.
type feishuDocRawResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Content string `json:"content"` // Docx block JSON string
	} `json:"data"`
}

// feishuDocContentResp represents the content of a new-style docx document.
type feishuDocContentResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Blocks []feishuBlock `json:"blocks"`
	} `json:"data"`
}

type feishuBlock struct {
	BlockID   string   `json:"block_id"`
	BlockType int      `json:"block_type"`
	ParentID  string   `json:"parent_id"`
	Children  []string `json:"children"`
	Text      *feishuText `json:"text,omitempty"`
	// Other block types
	Table *feishuBlockTable `json:"table,omitempty"`
}

type feishuText struct {
	Elements []feishuTextElement `json:"elements"`
	Style    *feishuBlockStyle   `json:"style,omitempty"`
}

type feishuTextElement struct {
	TextRun *feishuTextRun `json:"text_run,omitempty"`
	MentionUser *feishuMentionUser `json:"mention_user,omitempty"`
	MentionDoc  *feishuMentionDoc  `json:"mention_doc,omitempty"`
}

type feishuTextRun struct {
	Content string           `json:"content"`
	Style   *feishuTextStyle `json:"text_element_style,omitempty"`
}

type feishuTextStyle struct {
	Bold     bool   `json:"bold"`
	Link     string `json:"link,omitempty"`
}

type feishuBlockStyle struct {
	Align           int  `json:"align"`
	BackgroundColor int  `json:"background_color,omitempty"`
}

type feishuBlockTable struct {
	// Simplified table handling
}

type feishuMentionUser struct {
	UserID string `json:"user_id"`
}

type feishuMentionDoc struct {
	DocToken string `json:"token"`
}

// FeishuConnector syncs documents from a Feishu/Lark knowledge space into the RAG system.
type FeishuConnector struct {
	config    *FeishuSourceConfig
	kbID      string
	accessToken string
	httpClient  *http.Client
}

const feishuBaseURL = "https://open.feishu.cn/open-apis"

// NewFeishuConnector creates a new connector for syncing Feishu docs.
func NewFeishuConnector(kbID string, sourceConfigJSON string) (*FeishuConnector, error) {
	var cfg FeishuSourceConfig
	if err := json.Unmarshal([]byte(sourceConfigJSON), &cfg); err != nil {
		return nil, fmt.Errorf("parse feishu source_config: %w", err)
	}

	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil, fmt.Errorf("feishu app_id and app_secret are required")
	}
	if cfg.WikiToken == "" {
		return nil, fmt.Errorf("feishu wiki_token is required")
	}

	if cfg.WikiType == "" {
		cfg.WikiType = "wiki"
	}
	if cfg.NodeToken == "" {
		cfg.NodeToken = cfg.WikiToken
	}

	return &FeishuConnector{
		config:   &cfg,
		kbID:     kbID,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// getAccessToken obtains a tenant access token from Feishu Open API.
func (f *FeishuConnector) getAccessToken(ctx context.Context) (string, error) {
	body := map[string]string{
		"app_id":     f.config.AppID,
		"app_secret": f.config.AppSecret,
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST",
		feishuBaseURL+"/auth/v3/tenant_access_token/internal",
		bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get tenant token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int64  `json:"expire"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token resp: %w", err)
	}

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("feishu auth error: %d %s", tokenResp.Code, tokenResp.Msg)
	}

	f.accessToken = tokenResp.TenantAccessToken
	return f.accessToken, nil
}

// feishuAPI is a helper to make authenticated Feishu API calls.
func (f *FeishuConnector) feishuAPI(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := feishuBaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.accessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return f.httpClient.Do(req)
}

// SyncDocument is the core sync method. It recursively syncs all documents from the wiki.
// processDocCallback is called for each document to handle indexing.
func (f *FeishuConnector) SyncDocument(ctx context.Context, processDocCallback func(docName, docContent string) error) error {
	if _, err := f.getAccessToken(ctx); err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	// If wiki_type is "drive", use cloud space API instead
	if f.config.WikiType == "drive" {
		return f.syncDriveRecursive(ctx, f.config.NodeToken, processDocCallback)
	}

	// For wiki type, use wiki node API
	return f.syncWikiRecursive(ctx, f.config.NodeToken, processDocCallback)
}

// syncWikiRecursive recursively syncs wiki space nodes.
func (f *FeishuConnector) syncWikiRecursive(ctx context.Context, nodeToken string, processDocCallback func(docName, docContent string) error) error {
	// First, get node info to determine what type it is
	node, err := f.getWikiNode(ctx, nodeToken)
	if err != nil {
		return fmt.Errorf("get node %s: %w", nodeToken, err)
	}

	objType := node.Data.Node.ObjType
	title := node.Data.Node.Title
	hasChild := node.Data.Node.HasChild

	switch objType {
	case "doc", "docx":
		content, err := f.fetchDocContent(ctx, nodeToken, objType)
		if err != nil {
			return fmt.Errorf("fetch doc %s: %w", title, err)
		}
		return processDocCallback(title, content)

	case "sheet":
		// For sheets, try to export as text
		content, err := f.fetchSheetContent(ctx, nodeToken)
		if err != nil {
			return fmt.Errorf("fetch sheet %s: %w", title, err)
		}
		return processDocCallback(title, content)

	case "bitable":
		// Skip bitables for now - too complex to textify
		return nil

	case "folder", "mindnote":
		// Recurse into children
		if hasChild {
			return f.listAndSyncChildren(ctx, nodeToken, processDocCallback)
		}
		return nil

	case "shortcut":
		// Follow shortcuts
		// Not directly queryable, skip
		return nil

	default:
		// Try as doc anyway
		content, err := f.fetchDocContent(ctx, nodeToken, "docx")
		if err != nil {
			// Not a readable document, skip
			return nil
		}
		return processDocCallback(title, content)
	}
}

// syncDriveRecursive recursively syncs cloud space (drive) files.
func (f *FeishuConnector) syncDriveRecursive(ctx context.Context, folderToken string, processDocCallback func(docName, docContent string) error) error {
	return f.listDriveChildren(ctx, folderToken, processDocCallback)
}

// getWikiNode fetches wiki node information.
func (f *FeishuConnector) getWikiNode(ctx context.Context, nodeToken string) (*feishuNodeInfoResp, error) {
	resp, err := f.feishuAPI(ctx, "GET",
		fmt.Sprintf("/wiki/v2/spaces/%s/nodes/%s", f.config.WikiToken, nodeToken), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result feishuNodeInfoResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode node resp: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("feishu api error: %d %s", result.Code, result.Msg)
	}
	return &result, nil
}

// listAndSyncChildren lists children of a wiki node and syncs each.
func (f *FeishuConnector) listAndSyncChildren(ctx context.Context, parentToken string, processDocCallback func(docName, docContent string) error) error {
	pageToken := ""
	for {
		path := fmt.Sprintf("/wiki/v2/spaces/%s/nodes/%s/children?page_size=50", f.config.WikiToken, parentToken)
		if pageToken != "" {
			path += "&page_token=" + pageToken
		}

		resp, err := f.feishuAPI(ctx, "GET", path, nil)
		if err != nil {
			return err
		}

		var children feishuChildrenResp
		if err := json.NewDecoder(resp.Body).Decode(&children); err != nil {
			resp.Body.Close()
			return fmt.Errorf("decode children: %w", err)
		}
		resp.Body.Close()

		if children.Code != 0 {
			return fmt.Errorf("list children error: %d %s", children.Code, children.Msg)
		}

		for _, item := range children.Data.Items {
			if err := f.syncWikiRecursive(ctx, item.Token, processDocCallback); err != nil {
				// Log and continue on individual doc errors
				continue
			}
		}

		if !children.Data.HasMore {
			break
		}
		pageToken = children.Data.PageToken
	}
	return nil
}

// listDriveChildren lists children of a cloud space folder and syncs each.
func (f *FeishuConnector) listDriveChildren(ctx context.Context, folderToken string, processDocCallback func(docName, docContent string) error) error {
	pageToken := ""
	for {
		path := fmt.Sprintf("/drive/v1/files/%s/children?page_size=50&types=doc,docx,sheet", folderToken)
		if pageToken != "" {
			path += "&page_token=" + pageToken
		}

		resp, err := f.feishuAPI(ctx, "GET", path, nil)
		if err != nil {
			return err
		}

		var children feishuChildrenResp
		if err := json.NewDecoder(resp.Body).Decode(&children); err != nil {
			resp.Body.Close()
			return fmt.Errorf("decode children: %w", err)
		}
		resp.Body.Close()

		if children.Code != 0 {
			return fmt.Errorf("list drive children error: %d %s", children.Code, children.Msg)
		}

		for _, item := range children.Data.Items {
			content, _ := f.fetchDocContent(ctx, item.Token, item.ObjType)
			if content != "" {
				if err := processDocCallback(item.Title, content); err != nil {
					continue
				}
			}
			if item.HasChild {
				_ = f.listDriveChildren(ctx, item.Token, processDocCallback)
			}
		}

		if !children.Data.HasMore {
			break
		}
		pageToken = children.Data.PageToken
	}
	return nil
}

// fetchDocContent fetches and converts a Feishu document to plain text.
func (f *FeishuConnector) fetchDocContent(ctx context.Context, docToken, objType string) (string, error) {
	switch objType {
	case "docx":
		return f.fetchDocxContent(ctx, docToken)
	case "doc":
		return f.fetchDocRawContent(ctx, docToken)
	default:
		// Try docx first, then doc
		content, err := f.fetchDocxContent(ctx, docToken)
		if err == nil {
			return content, nil
		}
		return f.fetchDocRawContent(ctx, docToken)
	}
}

// fetchDocxContent fetches new-style docx document content.
func (f *FeishuConnector) fetchDocxContent(ctx context.Context, docToken string) (string, error) {
	resp, err := f.feishuAPI(ctx, "GET",
		fmt.Sprintf("/docx/v1/documents/%s/blocks?page_size=500", docToken), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result feishuDocContentResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode docx: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("feishu docx api: %d %s", result.Code, result.Msg)
	}

	// Convert blocks to text
	var text string
	for _, block := range result.Data.Blocks {
		text += renderBlock(block) + "\n"
	}
	return text, nil
}

// fetchDocRawContent fetches old-style doc content.
func (f *FeishuConnector) fetchDocRawContent(ctx context.Context, docToken string) (string, error) {
	resp, err := f.feishuAPI(ctx, "GET",
		fmt.Sprintf("/doc/v2/meta/%s/content", docToken), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result feishuDocRawResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode doc: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("feishu doc api: %d %s", result.Code, result.Msg)
	}
	return result.Data.Content, nil
}

// fetchSheetContent exports a Feishu sheet as plain text (tsv).
func (f *FeishuConnector) fetchSheetContent(ctx context.Context, sheetToken string) (string, error) {
	// Use spreadsheet API to get sheet data
	resp, err := f.feishuAPI(ctx, "GET",
		fmt.Sprintf("/sheets/v2/spreadsheets/%s/metainfo", sheetToken), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// For simplicity, fetch range values
	var meta struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Properties struct {
				Title       string `json:"title"`
				SheetCount  int    `json:"sheet_count"`
			} `json:"properties"`
			Sheets []struct {
				SheetID string `json:"sheet_id"`
				Title   string `json:"title"`
			} `json:"sheets"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", fmt.Errorf("decode sheet meta: %w", err)
	}
	if meta.Code != 0 {
		return "", fmt.Errorf("sheet meta api: %d %s", meta.Code, meta.Msg)
	}

	var allContent string
	for _, sheet := range meta.Data.Sheets {
		sheetContent, err := f.fetchSheetRange(ctx, sheetToken, sheet.SheetID, "A1:ZZ1000")
		if err != nil {
			continue
		}
		allContent += fmt.Sprintf("=== %s ===\n%s\n\n", sheet.Title, sheetContent)
	}
	return allContent, nil
}

// fetchSheetRange fetches cell values from a sheet range.
func (f *FeishuConnector) fetchSheetRange(ctx context.Context, sheetToken, sheetID, rangeStr string) (string, error) {
	resp, err := f.feishuAPI(ctx, "GET",
		fmt.Sprintf("/sheets/v2/spreadsheets/%s/values/%s?value_render_option=ToString", sheetToken, rangeStr), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ValueRange struct {
				Values [][]interface{} `json:"values"`
			} `json:"value_range"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Code != 0 {
		return "", fmt.Errorf("sheet values api: %d %s", result.Code, result.Msg)
	}

	var text string
	for _, row := range result.Data.ValueRange.Values {
		var cells []string
		for _, cell := range row {
			if cell != nil {
				cells = append(cells, fmt.Sprintf("%v", cell))
			}
		}
		text += fmt.Sprintf("%s\n", joinTab(cells))
	}
	return text, nil
}

// renderBlock converts a Feishu docx block to plain text.
func renderBlock(block feishuBlock) string {
	if block.Text == nil {
		return ""
	}

	var content string
	for _, elem := range block.Text.Elements {
		if elem.TextRun != nil {
			content += elem.TextRun.Content
		} else if elem.MentionUser != nil {
			content += fmt.Sprintf("@%s", elem.MentionUser.UserID)
		} else if elem.MentionDoc != nil {
			content += fmt.Sprintf("[文档:%s]", elem.MentionDoc.DocToken)
		}
	}

	// Add block-level formatting based on block type
	switch block.BlockType {
	case 3: // Heading 1
		content = "# " + content
	case 4: // Heading 2
		content = "## " + content
	case 5: // Heading 3
		content = "### " + content
	case 6, 7, 8, 9, 10, 11: // Heading 4-9
		content = "#### " + content
	case 12: // Bullet list
		content = "- " + content
	case 13: // Ordered list
		content = "1. " + content
	case 14: // Code block
		content = "```\n" + content + "\n```"
	case 15: // Quote
		content = "> " + content
	case 22: // Task
		content = "- [ ] " + content
	}

	return content
}

// joinTab joins strings with tab as separator.
func joinTab(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += "\t" + items[i]
	}
	return result
}

// Ensure implements a simple join with tab separator
var _ = joinTab

// SyncDocumentsWithPipeline is the main entry point used by the external sync trigger.
// It creates a FeishuConnector, syncs documents, and processes each through the document pipeline.
func SyncFeishuDocuments(ctx context.Context, kb *dao.KnowledgeBaseRow, createDocFn func(docName, contentType string, data []byte) (string, error)) error {
	connector, err := NewFeishuConnector(kb.KbID, kb.SourceConfig)
	if err != nil {
		return fmt.Errorf("create feishu connector: %w", err)
	}

	return connector.SyncDocument(ctx, func(docName, content string) error {
		if content == "" || docName == "" {
			return nil
		}
		// Create a document in the RAG system as text/plain
		_, err := createDocFn(docName, "text/plain", []byte(content))
		return err
	})
}