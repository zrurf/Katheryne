package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HugeGraphDao wraps the HugeGraph REST API for graph operations.
type HugeGraphDao struct {
	baseURL    string
	graph      string
	httpClient *http.Client
}

// NewHugeGraphDao creates a new HugeGraph DAO.
func NewHugeGraphDao(baseURL, graph string) *HugeGraphDao {
	return &HugeGraphDao{
		baseURL: baseURL,
		graph:   graph,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EnsureSchema creates the graph schema if not exists (vertex labels, edge labels).
func (h *HugeGraphDao) EnsureSchema(ctx context.Context) error {
	// Create property keys
	props := []map[string]interface{}{
		{"name": "name", "data_type": "TEXT", "cardinality": "SINGLE"},
		{"name": "type", "data_type": "TEXT", "cardinality": "SINGLE"},
		{"name": "kb_id", "data_type": "TEXT", "cardinality": "SINGLE"},
		{"name": "doc_id", "data_type": "TEXT", "cardinality": "SINGLE"},
		{"name": "chunk_id", "data_type": "TEXT", "cardinality": "SINGLE"},
		{"name": "description", "data_type": "TEXT", "cardinality": "SINGLE"},
		{"name": "created_at", "data_type": "LONG", "cardinality": "SINGLE"},
	}
	for _, p := range props {
		_ = h.createPropertyKey(ctx, p) // ignore if already exists
	}

	// Create vertex labels
	vertices := []map[string]interface{}{
		{"name": "entity",
			"properties":        []string{"name", "type", "kb_id", "description", "created_at"},
			"primary_keys":      []string{"name"},
			"nullable_keys":     []string{"description"},
			"enable_label_index": true},
		{"name": "document",
			"properties":        []string{"name", "kb_id", "doc_id", "created_at"},
			"primary_keys":      []string{"doc_id"},
			"enable_label_index": true},
		{"name": "chunk",
			"properties":        []string{"name", "kb_id", "doc_id", "chunk_id", "created_at"},
			"primary_keys":      []string{"chunk_id"},
			"enable_label_index": true},
	}
	for _, v := range vertices {
		_ = h.createVertexLabel(ctx, v)
	}

	// Create edge labels
	edges := []map[string]interface{}{
		{"name": "contains",
			"source_label":      "document",
			"target_label":      "chunk",
			"properties":        []string{"kb_id", "created_at"},
			"nullable_keys":     []string{},
			"enable_label_index": true},
		{"name": "mentions",
			"source_label":      "chunk",
			"target_label":      "entity",
			"properties":        []string{"kb_id", "created_at"},
			"nullable_keys":     []string{},
			"enable_label_index": true},
		{"name": "relates_to",
			"source_label":      "entity",
			"target_label":      "entity",
			"properties":        []string{"relation_type", "chunk_id", "kb_id", "created_at"},
			"nullable_keys":     []string{},
			"enable_label_index": true},
	}
	for _, e := range edges {
		_ = h.createEdgeLabel(ctx, e)
	}
	return nil
}

// AddEntity adds an entity vertex to the graph.
func (h *HugeGraphDao) AddEntity(ctx context.Context, entityID, name, entityType, kbID, description string) error {
	vertex := map[string]interface{}{
		"label": "entity",
		"id":    h.makeVertexID("entity", entityID),
		"properties": map[string]interface{}{
			"name":        name,
			"type":        entityType,
			"kb_id":       kbID,
			"description": description,
			"created_at":  time.Now().UnixMilli(),
		},
	}
	return h.upsertVertex(ctx, vertex)
}

// AddRelation adds a relation edge between two entities.
func (h *HugeGraphDao) AddRelation(ctx context.Context, sourceEntityID, targetEntityID, relationType, chunkID, kbID string) error {
	edge := map[string]interface{}{
		"label":       "relates_to",
		"outV":        h.makeVertexID("entity", sourceEntityID),
		"inV":         h.makeVertexID("entity", targetEntityID),
		"outVLabel":   "entity",
		"inVLabel":    "entity",
		"properties": map[string]interface{}{
			"relation_type": relationType,
			"chunk_id":      chunkID,
			"kb_id":         kbID,
			"created_at":    time.Now().UnixMilli(),
		},
	}
	return h.upsertEdge(ctx, edge)
}

// AddDocumentVertex adds a document vertex.
func (h *HugeGraphDao) AddDocumentVertex(ctx context.Context, docID, name, kbID string) error {
	vertex := map[string]interface{}{
		"label": "document",
		"id":    h.makeVertexID("document", docID),
		"properties": map[string]interface{}{
			"name":       name,
			"kb_id":      kbID,
			"doc_id":     docID,
			"created_at": time.Now().UnixMilli(),
		},
	}
	return h.upsertVertex(ctx, vertex)
}

// AddChunkVertex adds a chunk vertex and links it to document.
func (h *HugeGraphDao) AddChunkVertex(ctx context.Context, chunkID, docID, kbID string, chunkIndex int) error {
	vertex := map[string]interface{}{
		"label": "chunk",
		"id":    h.makeVertexID("chunk", chunkID),
		"properties": map[string]interface{}{
			"name":       fmt.Sprintf("chunk-%d", chunkIndex),
			"kb_id":      kbID,
			"doc_id":     docID,
			"chunk_id":   chunkID,
			"created_at": time.Now().UnixMilli(),
		},
	}
	if err := h.upsertVertex(ctx, vertex); err != nil {
		return err
	}
	// Link document -> chunk
	edge := map[string]interface{}{
		"label":     "contains",
		"outV":      h.makeVertexID("document", docID),
		"inV":       h.makeVertexID("chunk", chunkID),
		"outVLabel": "document",
		"inVLabel":  "chunk",
		"properties": map[string]interface{}{
			"kb_id":      kbID,
			"created_at": time.Now().UnixMilli(),
		},
	}
	return h.upsertEdge(ctx, edge)
}

// LinkChunkToEntity links a chunk to an entity (mentions).
func (h *HugeGraphDao) LinkChunkToEntity(ctx context.Context, chunkID, entityID, kbID string) error {
	edge := map[string]interface{}{
		"label":     "mentions",
		"outV":      h.makeVertexID("chunk", chunkID),
		"inV":       h.makeVertexID("entity", entityID),
		"outVLabel": "chunk",
		"inVLabel":  "entity",
		"properties": map[string]interface{}{
			"kb_id":      kbID,
			"created_at": time.Now().UnixMilli(),
		},
	}
	return h.upsertEdge(ctx, edge)
}

// GetSubGraph traverses the graph around a central entity.
func (h *HugeGraphDao) GetSubGraph(ctx context.Context, entityID string, depth, limit int) ([]map[string]interface{}, []map[string]interface{}, error) {
	gremlin := fmt.Sprintf(
		`g.V("%s").repeat(__.bothE().dedup().as("e").bothV().dedup()).times(%d).emit().path().limit(%d)`,
		h.makeVertexID("entity", entityID), depth, limit,
	)
	return h.executeGremlin(ctx, gremlin)
}

// DeleteGraphByKB removes all vertices and edges for a knowledge base.
func (h *HugeGraphDao) DeleteGraphByKB(ctx context.Context, kbID string) error {
	gremlin := fmt.Sprintf(`g.V().has("kb_id", "%s").drop()`, kbID)
	_, err := h.doRequest(ctx, "GET", fmt.Sprintf("/gremlin?gremlin=%s", gremlin), nil)
	return err
}

// GetKBEntityCount returns entity count for a KB.
func (h *HugeGraphDao) GetKBEntityCount(ctx context.Context, kbID string) (int64, error) {
	gremlin := fmt.Sprintf(`g.V().has("kb_id", "%s").hasLabel("entity").count()`, kbID)
	resp, err := h.doRequest(ctx, "GET", fmt.Sprintf("/gremlin?gremlin=%s", gremlin), nil)
	if err != nil {
		return 0, err
	}
	var result []int64
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, err
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return 0, nil
}

// GetKBRelationCount returns relation count for a KB.
func (h *HugeGraphDao) GetKBRelationCount(ctx context.Context, kbID string) (int64, error) {
	gremlin := fmt.Sprintf(`g.E().has("kb_id", "%s").count()`, kbID)
	resp, err := h.doRequest(ctx, "GET", fmt.Sprintf("/gremlin?gremlin=%s", gremlin), nil)
	if err != nil {
		return 0, err
	}
	var result []int64
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, err
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return 0, nil
}

func (h *HugeGraphDao) makeVertexID(label, id string) string {
	return fmt.Sprintf("%s:%s", label, id)
}

func (h *HugeGraphDao) upsertVertex(ctx context.Context, vertex map[string]interface{}) error {
	body, _ := json.Marshal(vertex)
	_, err := h.doRequest(ctx, "POST", fmt.Sprintf("/graph/%s/graph/vertices", h.graph), body)
	return err
}

func (h *HugeGraphDao) upsertEdge(ctx context.Context, edge map[string]interface{}) error {
	body, _ := json.Marshal(edge)
	_, err := h.doRequest(ctx, "POST", fmt.Sprintf("/graph/%s/graph/edges", h.graph), body)
	return err
}

func (h *HugeGraphDao) createPropertyKey(ctx context.Context, props map[string]interface{}) error {
	body, _ := json.Marshal(props)
	_, err := h.doRequest(ctx, "POST", fmt.Sprintf("/graph/%s/schema/propertykeys", h.graph), body)
	return err
}

func (h *HugeGraphDao) createVertexLabel(ctx context.Context, vl map[string]interface{}) error {
	body, _ := json.Marshal(vl)
	_, err := h.doRequest(ctx, "POST", fmt.Sprintf("/graph/%s/schema/vertexlabels", h.graph), body)
	return err
}

func (h *HugeGraphDao) createEdgeLabel(ctx context.Context, el map[string]interface{}) error {
	body, _ := json.Marshal(el)
	_, err := h.doRequest(ctx, "POST", fmt.Sprintf("/graph/%s/schema/edgelabels", h.graph), body)
	return err
}

func (h *HugeGraphDao) executeGremlin(ctx context.Context, gremlin string) ([]map[string]interface{}, []map[string]interface{}, error) {
	resp, err := h.doRequest(ctx, "GET", fmt.Sprintf("/gremlin?gremlin=%s", gremlin), nil)
	if err != nil {
		return nil, nil, err
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(resp, &raw); err != nil {
		return nil, nil, err
	}

	// Parse vertices and edges from path results
	var vertices, edges []map[string]interface{}
	for _, item := range raw {
		if objects, ok := item["objects"].([]interface{}); ok {
			for _, obj := range objects {
				m, ok := obj.(map[string]interface{})
				if !ok {
					continue
				}
				if m["label"] == "vertex" {
					vertices = append(vertices, m)
				} else if m["label"] == "edge" {
					edges = append(edges, m)
				}
			}
		}
	}
	return vertices, edges, nil
}

func (h *HugeGraphDao) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	url := h.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Accept 2xx and 409 (already exists)
	if resp.StatusCode >= 400 && resp.StatusCode != 409 {
		return nil, fmt.Errorf("hugegraph api error [%d]: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}