package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeromicro/go-zero/core/logx"
)

// MemoryRow represents a memory entry in PostgreSQL.
type MemoryRow struct {
	MemoryID       string            `db:"memory_id"`
	TenantID       string            `db:"tenant_id"`
	TenantType     string            `db:"tenant_type"`
	MemoryType     int32             `db:"memory_type"`
	Content        string            `db:"content"`
	Importance     float64           `db:"importance"`
	Entities       []string          `db:"entities"`
	Metadata       map[string]string `db:"metadata"`
	AccessCount    int32             `db:"access_count"`
	ExpiresAt      *time.Time        `db:"expires_at"`
	CreatedAt      time.Time         `db:"created_at"`
	UpdatedAt      time.Time         `db:"updated_at"`
	LastAccessedAt time.Time         `db:"last_accessed_at"`
}

// PostgresDao manages PostgreSQL operations for the memory service.
type PostgresDao struct {
	pool *pgxpool.Pool
}

// NewPostgresDao creates a new PostgreSQL DAO and ensures the schema.
func NewPostgresDao(host string, port int, user, password, dbname string) (*PostgresDao, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}
	config.MaxConns = 20

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect pg: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}

	d := &PostgresDao{pool: pool}
	if err := d.ensureSchema(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	return d, nil
}

func (d *PostgresDao) Close() {
	d.pool.Close()
}

func (d *PostgresDao) ensureSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		memory_id        TEXT PRIMARY KEY,
		tenant_id        TEXT NOT NULL,
		tenant_type      TEXT NOT NULL,
		memory_type      INT NOT NULL DEFAULT 0,
		content          TEXT NOT NULL,
		importance       DOUBLE PRECISION NOT NULL DEFAULT 0.5,
		entities         TEXT[] DEFAULT '{}',
		metadata         JSONB DEFAULT '{}',
		access_count     INT NOT NULL DEFAULT 0,
		expires_at       TIMESTAMPTZ,
		created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_memories_tenant ON memories(tenant_id, tenant_type);
	CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(memory_type);
	CREATE INDEX IF NOT EXISTS idx_memories_importance ON memories(importance);
	CREATE INDEX IF NOT EXISTS idx_memories_expires ON memories(expires_at) WHERE expires_at IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_memories_fts ON memories USING GIN (to_tsvector('simple', content));

	CREATE TABLE IF NOT EXISTS memory_graph (
		edge_id       TEXT PRIMARY KEY,
		tenant_id     TEXT NOT NULL,
		tenant_type   TEXT NOT NULL,
		source_entity TEXT NOT NULL,
		target_entity TEXT NOT NULL,
		relation_type TEXT NOT NULL,
		weight        DOUBLE PRECISION NOT NULL DEFAULT 1.0,
		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_memory_graph_tenant ON memory_graph(tenant_id, tenant_type);
	CREATE INDEX IF NOT EXISTS idx_memory_graph_source ON memory_graph(source_entity);
	CREATE INDEX IF NOT EXISTS idx_memory_graph_target ON memory_graph(target_entity);
	`
	_, err := d.pool.Exec(ctx, schema)
	return err
}

// InsertMemory inserts a new memory record.
func (d *PostgresDao) InsertMemory(ctx context.Context, row *MemoryRow) error {
	var expiresAt interface{}
	if row.ExpiresAt != nil {
		expiresAt = *row.ExpiresAt
	}

	_, err := d.pool.Exec(ctx, `
		INSERT INTO memories (memory_id, tenant_id, tenant_type, memory_type, content, importance, entities, metadata, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, row.MemoryID, row.TenantID, row.TenantType, row.MemoryType, row.Content, row.Importance,
		row.Entities, row.Metadata, expiresAt)
	return err
}

// UpdateMemory updates an existing memory.
func (d *PostgresDao) UpdateMemory(ctx context.Context, memoryID, content string, importance float64, metadata map[string]string) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE memories SET content=$2, importance=$3, metadata=$4, updated_at=NOW()
		WHERE memory_id=$1
	`, memoryID, content, importance, metadata)
	return err
}

// DeleteMemory deletes a memory by ID.
func (d *PostgresDao) DeleteMemory(ctx context.Context, memoryID string) error {
	_, err := d.pool.Exec(ctx, `DELETE FROM memories WHERE memory_id=$1`, memoryID)
	return err
}

// ClearMemories deletes all memories for a tenant.
func (d *PostgresDao) ClearMemories(ctx context.Context, tenantID, tenantType string) (int64, error) {
	tag, err := d.pool.Exec(ctx, `DELETE FROM memories WHERE tenant_id=$1 AND tenant_type=$2`, tenantID, tenantType)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// GetMemory returns a single memory by ID.
func (d *PostgresDao) GetMemory(ctx context.Context, memoryID string) (*MemoryRow, error) {
	row := &MemoryRow{}
	var expiresAt *time.Time
	err := d.pool.QueryRow(ctx, `
		SELECT memory_id, tenant_id, tenant_type, memory_type, content, importance,
		       entities, metadata, access_count, expires_at, created_at, updated_at, last_accessed_at
		FROM memories WHERE memory_id=$1
	`, memoryID).Scan(&row.MemoryID, &row.TenantID, &row.TenantType, &row.MemoryType,
		&row.Content, &row.Importance, &row.Entities, &row.Metadata,
		&row.AccessCount, &expiresAt, &row.CreatedAt, &row.UpdatedAt, &row.LastAccessedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if expiresAt != nil {
		row.ExpiresAt = expiresAt
	}
	return row, nil
}

// ListMemories returns memories for a tenant with optional type filter.
func (d *PostgresDao) ListMemories(ctx context.Context, tenantID, tenantType string, memoryType int32, page, size int32) ([]*MemoryRow, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// Count
	var total int64
	countSQL := `SELECT COUNT(*) FROM memories WHERE tenant_id=$1 AND tenant_type=$2`
	args := []interface{}{tenantID, tenantType}
	if memoryType > 0 {
		countSQL += ` AND memory_type=$3`
		args = append(args, memoryType)
	}
	if err := d.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Query
	querySQL := `SELECT memory_id, tenant_id, tenant_type, memory_type, content, importance,
	                    entities, metadata, access_count, expires_at, created_at, updated_at, last_accessed_at
	             FROM memories WHERE tenant_id=$1 AND tenant_type=$2`
	queryArgs := []interface{}{tenantID, tenantType}
	if memoryType > 0 {
		querySQL += ` AND memory_type=$3`
		queryArgs = append(queryArgs, memoryType)
	}
	querySQL += ` ORDER BY updated_at DESC LIMIT $` + fmt.Sprintf("%d", len(queryArgs)+1) +
		` OFFSET $` + fmt.Sprintf("%d", len(queryArgs)+2)
	queryArgs = append(queryArgs, size, offset)

	rows, err := d.pool.Query(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*MemoryRow
	for rows.Next() {
		row := &MemoryRow{}
		var expiresAt *time.Time
		if err := rows.Scan(&row.MemoryID, &row.TenantID, &row.TenantType, &row.MemoryType,
			&row.Content, &row.Importance, &row.Entities, &row.Metadata,
			&row.AccessCount, &expiresAt, &row.CreatedAt, &row.UpdatedAt, &row.LastAccessedAt); err != nil {
			return nil, 0, err
		}
		if expiresAt != nil {
			row.ExpiresAt = expiresAt
		}
		list = append(list, row)
	}
	return list, total, nil
}

// FullTextSearch performs keyword-based full-text search on memories.
func (d *PostgresDao) FullTextSearch(ctx context.Context, tenantID, tenantType, query string, topK int32, memoryTypes []int32) ([]*MemoryRow, error) {
	if query == "" {
		return nil, nil
	}

	args := []interface{}{tenantID, tenantType, strings.ReplaceAll(query, "'", "''")}
	baseSQL := `
		SELECT memory_id, tenant_id, tenant_type, memory_type, content, importance,
		       entities, metadata, access_count, expires_at, created_at, updated_at, last_accessed_at,
		       ts_rank(to_tsvector('simple', content), plainto_tsquery('simple', $3)) AS rank
		FROM memories
		WHERE tenant_id=$1 AND tenant_type=$2
		  AND to_tsvector('simple', content) @@ plainto_tsquery('simple', $3)`

	if len(memoryTypes) > 0 {
		placeholders := make([]string, len(memoryTypes))
		for i, mt := range memoryTypes {
			placeholders[i] = fmt.Sprintf("$%d", 4+i)
			args = append(args, mt)
		}
		baseSQL += ` AND memory_type IN (` + strings.Join(placeholders, ",") + `)`
	}

	baseSQL += ` ORDER BY rank DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, topK)

	rows, err := d.pool.Query(ctx, baseSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*MemoryRow
	for rows.Next() {
		row := &MemoryRow{}
		var rank float64
		var expiresAt *time.Time
		if err := rows.Scan(&row.MemoryID, &row.TenantID, &row.TenantType, &row.MemoryType,
			&row.Content, &row.Importance, &row.Entities, &row.Metadata,
			&row.AccessCount, &expiresAt, &row.CreatedAt, &row.UpdatedAt, &row.LastAccessedAt,
			&rank); err != nil {
			return nil, err
		}
		if expiresAt != nil {
			row.ExpiresAt = expiresAt
		}
		list = append(list, row)
	}
	return list, nil
}

// RecordAccess increments the access count and updates last_accessed_at.
func (d *PostgresDao) RecordAccess(ctx context.Context, memoryID string) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE memories SET access_count = access_count + 1, last_accessed_at = NOW()
		WHERE memory_id=$1
	`, memoryID)
	if err != nil {
		logx.Errorf("record access for %s: %v", memoryID, err)
	}
	return err
}

// DeleteExpired removes all expired memories. Returns count deleted.
func (d *PostgresDao) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := d.pool.Exec(ctx, `DELETE FROM memories WHERE expires_at IS NOT NULL AND expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// ========== Graph Operations ==========

// UpsertGraphEdge inserts or updates a graph edge.
func (d *PostgresDao) UpsertGraphEdge(ctx context.Context, edgeID, tenantID, tenantType, source, target, relationType string, weight float64) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO memory_graph (edge_id, tenant_id, tenant_type, source_entity, target_entity, relation_type, weight)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (edge_id) DO UPDATE SET weight=$7, created_at=NOW()
	`, edgeID, tenantID, tenantType, source, target, relationType, weight)
	return err
}

// QueryGraphTraverse performs BFS traversal from source entities.
func (d *PostgresDao) QueryGraphTraverse(ctx context.Context, tenantID, tenantType string, entities []string, depth, limit int32) ([]GraphNode, []GraphEdge, error) {
	if len(entities) == 0 {
		return nil, nil, nil
	}

	// Build a recursive CTE for graph traversal
	querySQL := `
		WITH RECURSIVE graph_walk AS (
			SELECT source_entity AS nid, target_entity, relation_type, weight, 1 AS d
			FROM memory_graph
			WHERE tenant_id=$1 AND tenant_type=$2 AND source_entity = ANY($3)
			UNION
			SELECT gw.target_entity AS nid, mg.target_entity, mg.relation_type, mg.weight, gw.d + 1
			FROM graph_walk gw
			JOIN memory_graph mg ON mg.source_entity = gw.target_entity
				AND mg.tenant_id=$1 AND mg.tenant_type=$2
			WHERE gw.d < $4
		)
		SELECT DISTINCT nid, target_entity, relation_type, weight, d FROM graph_walk LIMIT $5
	`

	rows, err := d.pool.Query(ctx, querySQL, tenantID, tenantType, entities, depth, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	seenNodes := make(map[string]bool)
	var nodes []GraphNode
	var edges []GraphEdge

	for rows.Next() {
		var nid, target, relType string
		var weight float64
		var dist int32
		if err := rows.Scan(&nid, &target, &relType, &weight, &dist); err != nil {
			return nil, nil, err
		}
		if !seenNodes[nid] {
			seenNodes[nid] = true
			nodes = append(nodes, GraphNode{EntityID: nid, Name: nid, EntityType: "entity"})
		}
		if !seenNodes[target] {
			seenNodes[target] = true
			nodes = append(nodes, GraphNode{EntityID: target, Name: target, EntityType: "entity"})
		}
		edges = append(edges, GraphEdge{
			EdgeID:       fmt.Sprintf("%s-%s-%s", nid, relType, target),
			Source:       nid,
			Target:       target,
			RelationType: relType,
			Weight:       weight,
		})
	}

	return nodes, edges, nil
}

// GraphNode and GraphEdge types for the DAO layer
type GraphNode struct {
	EntityID   string
	Name       string
	EntityType string
}

type GraphEdge struct {
	EdgeID       string
	Source       string
	Target       string
	RelationType string
	Weight       float64
}
