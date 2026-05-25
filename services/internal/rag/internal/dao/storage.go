package dao

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// KnowledgeBaseRow represents a knowledge base in PostgreSQL.
type KnowledgeBaseRow struct {
	KbID         string    `db:"kb_id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	OwnerUID     int64     `db:"owner_uid"`
	SourceType   string    `db:"source_type"`
	SourceConfig string    `db:"source_config"`
	Status       string    `db:"status"`
	DocCount     int64     `db:"doc_count"`
	ChunkCount   int64     `db:"chunk_count"`
	TotalSize    int64     `db:"total_size"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// DocumentRow represents a document in PostgreSQL.
type DocumentRow struct {
	DocID       string    `db:"doc_id"`
	KbID        string    `db:"kb_id"`
	FileName    string    `db:"file_name"`
	ContentType string    `db:"content_type"`
	FileSize    int64     `db:"file_size"`
	OssIndex    string    `db:"oss_index"`
	Status      string    `db:"status"`
	ChunkCount  int64     `db:"chunk_count"`
	ErrorMsg    string    `db:"error_msg"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// KBAuthorizationRow represents a KB authorization record.
type KBAuthorizationRow struct {
	KbID       string    `db:"kb_id"`
	BotID      int64     `db:"bot_id"`
	ConvID     int64     `db:"conv_id"`
	Permission string    `db:"permission"`
	GrantedAt  time.Time `db:"granted_at"`
}

// StorageDao handles PostgreSQL metadata operations.
type StorageDao struct {
	pool *pgxpool.Pool
}

func NewStorageDao(pool *pgxpool.Pool) *StorageDao {
	return &StorageDao{pool: pool}
}

// ===== Knowledge Base =====

func (s *StorageDao) CreateKB(ctx context.Context, kbID, name, description, sourceType, sourceConfig string, ownerUID int64) error {
	if sourceType == "" {
		sourceType = "PLATFORM"
	}
	if sourceConfig == "" {
		sourceConfig = "{}"
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO kb (kb_id, name, description, owner_uid, source_type, source_config) VALUES ($1, $2, $3, $4, $5, $6::jsonb)`,
		kbID, name, description, ownerUID, sourceType, sourceConfig,
	)
	return err
}

func (s *StorageDao) ListKBs(ctx context.Context, ownerUID int64, page, size int32) ([]KnowledgeBaseRow, int64, error) {
	var total int64
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kb WHERE owner_uid=$1 AND status='ACTIVE'`, ownerUID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := s.pool.Query(ctx,
		`SELECT kb_id, name, description, owner_uid,
		        COALESCE(source_type,'PLATFORM'), COALESCE(source_config::text,'{}'),
		        status, doc_count, chunk_count, total_size, created_at, updated_at
		 FROM kb WHERE owner_uid=$1 AND status='ACTIVE'
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		ownerUID, size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []KnowledgeBaseRow
	for rows.Next() {
		var r KnowledgeBaseRow
		if err := rows.Scan(&r.KbID, &r.Name, &r.Description, &r.OwnerUID,
			&r.SourceType, &r.SourceConfig,
			&r.Status, &r.DocCount, &r.ChunkCount, &r.TotalSize,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, nil
}

func (s *StorageDao) GetKB(ctx context.Context, kbID string) (*KnowledgeBaseRow, error) {
	var r KnowledgeBaseRow
	err := s.pool.QueryRow(ctx,
		`SELECT kb_id, name, description, owner_uid,
		        COALESCE(source_type,'PLATFORM'), COALESCE(source_config::text,'{}'),
		        status, doc_count, chunk_count, total_size, created_at, updated_at
		 FROM kb WHERE kb_id=$1`, kbID,
	).Scan(&r.KbID, &r.Name, &r.Description, &r.OwnerUID,
		&r.SourceType, &r.SourceConfig,
		&r.Status, &r.DocCount, &r.ChunkCount, &r.TotalSize,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *StorageDao) UpdateKB(ctx context.Context, kbID, name, description string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE kb SET name=$1, description=$2, updated_at=NOW() WHERE kb_id=$3`,
		name, description, kbID,
	)
	return err
}

func (s *StorageDao) DeleteKB(ctx context.Context, kbID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE kb SET status='DELETED', updated_at=NOW() WHERE kb_id=$1`, kbID,
	)
	return err
}

func (s *StorageDao) IncrementKBDocCount(ctx context.Context, kbID string, delta int, sizeDelta int64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE kb SET doc_count=doc_count+$1, total_size=total_size+$2, updated_at=NOW() WHERE kb_id=$3`,
		delta, sizeDelta, kbID,
	)
	return err
}

func (s *StorageDao) IncrementKBChunkCount(ctx context.Context, kbID string, delta int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE kb SET chunk_count=chunk_count+$1, updated_at=NOW() WHERE kb_id=$2`,
		delta, kbID,
	)
	return err
}

// ===== Documents =====

func (s *StorageDao) CreateDocument(ctx context.Context, docID, kbID, fileName, contentType string, fileSize int64, ossIndex string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO kb_document (doc_id, kb_id, file_name, content_type, file_size, oss_index, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'PROCESSING')`,
		docID, kbID, fileName, contentType, fileSize, ossIndex,
	)
	if err != nil {
		return err
	}
	return s.IncrementKBDocCount(ctx, kbID, 1, fileSize)
}

func (s *StorageDao) UpdateDocStatus(ctx context.Context, docID string, status, errorMsg string, chunkCount int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE kb_document SET status=$1, error_msg=$2, chunk_count=$3, updated_at=NOW() WHERE doc_id=$4`,
		status, errorMsg, chunkCount, docID,
	)
	return err
}

func (s *StorageDao) GetDocument(ctx context.Context, docID string) (*DocumentRow, error) {
	var r DocumentRow
	err := s.pool.QueryRow(ctx,
		`SELECT doc_id, kb_id, file_name, content_type, file_size, oss_index, status, chunk_count, error_msg, created_at, updated_at
		 FROM kb_document WHERE doc_id=$1`, docID,
	).Scan(&r.DocID, &r.KbID, &r.FileName, &r.ContentType, &r.FileSize, &r.OssIndex,
		&r.Status, &r.ChunkCount, &r.ErrorMsg, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *StorageDao) ListDocuments(ctx context.Context, kbID string, page, size int32) ([]DocumentRow, int64, error) {
	var total int64
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kb_document WHERE kb_id=$1 AND status!='DELETED'`, kbID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := s.pool.Query(ctx,
		`SELECT doc_id, kb_id, file_name, content_type, file_size, oss_index, status, chunk_count, error_msg, created_at, updated_at
		 FROM kb_document WHERE kb_id=$1 AND status!='DELETED'
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		kbID, size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []DocumentRow
	for rows.Next() {
		var r DocumentRow
		if err := rows.Scan(&r.DocID, &r.KbID, &r.FileName, &r.ContentType, &r.FileSize, &r.OssIndex,
			&r.Status, &r.ChunkCount, &r.ErrorMsg, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, nil
}

func (s *StorageDao) DeleteDocument(ctx context.Context, docID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE kb_document SET status='DELETED', updated_at=NOW() WHERE doc_id=$1`, docID,
	)
	return err
}

// ===== Authorization =====

func (s *StorageDao) GrantBotKBAccess(ctx context.Context, uid, botID, convID int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO kb_bot_access (uid, bot_id, conv_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		uid, botID, convID,
	)
	return err
}

func (s *StorageDao) RevokeBotKBAccess(ctx context.Context, uid, botID, convID int64) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM kb_bot_access WHERE uid=$1 AND bot_id=$2 AND conv_id=$3`,
		uid, botID, convID,
	)
	return err
}

func (s *StorageDao) CheckBotKBAccess(ctx context.Context, uid, botID, convID int64) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kb_bot_access WHERE uid=$1 AND bot_id=$2 AND conv_id=$3`,
		uid, botID, convID,
	).Scan(&count)
	return count > 0, err
}

func (s *StorageDao) AuthorizeKB(ctx context.Context, kbID string, botID, convID int64, permission string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO kb_auth (kb_id, bot_id, conv_id, permission) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (kb_id, bot_id, conv_id) DO UPDATE SET permission=$4, granted_at=NOW()`,
		kbID, botID, convID, permission,
	)
	return err
}

func (s *StorageDao) RevokeKBAuth(ctx context.Context, kbID string, botID, convID int64) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM kb_auth WHERE kb_id=$1 AND bot_id=$2 AND conv_id=$3`,
		kbID, botID, convID,
	)
	return err
}

func (s *StorageDao) CheckKBAuth(ctx context.Context, kbID string, botID, convID int64, permission string) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kb_auth WHERE kb_id=$1 AND bot_id=$2 AND conv_id=$3 AND permission=$4`,
		kbID, botID, convID, permission,
	).Scan(&count)
	return count > 0, err
}

func (s *StorageDao) ListKBAuths(ctx context.Context, uid int64, kbID string, botID int64) ([]KBAuthorizationRow, error) {
	query := `SELECT a.kb_id, a.bot_id, a.conv_id, a.permission, a.granted_at
		FROM kb_auth a JOIN kb k ON a.kb_id = k.kb_id
		WHERE k.owner_uid=$1`
	args := []interface{}{uid}
	argIdx := 2

	if kbID != "" {
		query += fmt.Sprintf(" AND a.kb_id=$%d", argIdx)
		args = append(args, kbID)
		argIdx++
	}
	if botID > 0 {
		query += fmt.Sprintf(" AND a.bot_id=$%d", argIdx)
		args = append(args, botID)
		argIdx++
	}

	query += " ORDER BY a.granted_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []KBAuthorizationRow
	for rows.Next() {
		var r KBAuthorizationRow
		if err := rows.Scan(&r.KbID, &r.BotID, &r.ConvID, &r.Permission, &r.GrantedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *StorageDao) ListBotKBs(ctx context.Context, botID, convID, uid int64) ([]KnowledgeBaseRow, error) {
	// Get KBs where bot has access AND user owns them
	rows, err := s.pool.Query(ctx,
		`SELECT DISTINCT k.kb_id, k.name, k.description, k.owner_uid,
		        COALESCE(k.source_type,'PLATFORM'), COALESCE(k.source_config::text,'{}'),
		        k.status, k.doc_count, k.chunk_count, k.total_size, k.created_at, k.updated_at
		 FROM kb k
		 JOIN kb_auth a ON k.kb_id = a.kb_id
		 WHERE a.bot_id=$1 AND a.conv_id=$2 AND k.owner_uid=$3 AND k.status='ACTIVE'`,
		botID, convID, uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []KnowledgeBaseRow
	for rows.Next() {
		var r KnowledgeBaseRow
		if err := rows.Scan(&r.KbID, &r.Name, &r.Description, &r.OwnerUID,
			&r.SourceType, &r.SourceConfig,
			&r.Status, &r.DocCount, &r.ChunkCount, &r.TotalSize,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// ChunkRow for document chunks metadata.
type ChunkRow struct {
	ChunkID    string
	DocID      string
	KbID       string
	Content    string
	ChunkIndex int32
	TokenCount int64
	Entities   string // JSON array
	CreatedAt  time.Time
}

func (s *StorageDao) GetDocChunks(ctx context.Context, docID string, page, size int32) ([]ChunkRow, int64, error) {
	var total int64
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kb_chunk WHERE doc_id=$1`, docID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := s.pool.Query(ctx,
		`SELECT chunk_id, doc_id, kb_id, content, chunk_index, token_count, entities, created_at
		 FROM kb_chunk WHERE doc_id=$1 ORDER BY chunk_index LIMIT $2 OFFSET $3`,
		docID, size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []ChunkRow
	for rows.Next() {
		var r ChunkRow
		if err := rows.Scan(&r.ChunkID, &r.DocID, &r.KbID, &r.Content, &r.ChunkIndex,
			&r.TokenCount, &r.Entities, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, nil
}

func (s *StorageDao) SaveChunk(ctx context.Context, chunkID, docID, kbID, content, entities string, chunkIndex int32, tokenCount int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO kb_chunk (chunk_id, doc_id, kb_id, content, chunk_index, token_count, entities)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		chunkID, docID, kbID, content, chunkIndex, tokenCount, entities,
	)
	return err
}

func (s *StorageDao) DeleteChunksByDoc(ctx context.Context, docID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM kb_chunk WHERE doc_id=$1`, docID)
	return err
}

// ===== External KB Sync =====

// ExternalSyncRow represents a kb_external_sync record.
type ExternalSyncRow struct {
	SyncID       string     `db:"sync_id"`
	KbID         string     `db:"kb_id"`
	SourceType   string     `db:"source_type"`
	SourceConfig string     `db:"source_config"`
	LastSyncedAt *time.Time `db:"last_synced_at"`
	SyncStatus   string     `db:"sync_status"`
	SyncError    string     `db:"sync_error"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

func (s *StorageDao) CreateExternalSync(ctx context.Context, kbID, sourceType, sourceConfig string) error {
	syncID := "sync_" + kbID
	_, err := s.pool.Exec(ctx,
		`INSERT INTO kb_external_sync (sync_id, kb_id, source_type, source_config, sync_status)
		 VALUES ($1, $2, $3, $4::jsonb, 'PENDING')
		 ON CONFLICT (sync_id) DO UPDATE SET source_config=$4::jsonb, sync_status='PENDING', updated_at=NOW()`,
		syncID, kbID, sourceType, sourceConfig,
	)
	return err
}

func (s *StorageDao) GetExternalSync(ctx context.Context, syncID string) (*ExternalSyncRow, error) {
	var r ExternalSyncRow
	err := s.pool.QueryRow(ctx,
		`SELECT sync_id, kb_id, source_type, COALESCE(source_config::text,'{}'),
		        last_synced_at, sync_status, COALESCE(sync_error,''),
		        created_at, updated_at
		 FROM kb_external_sync WHERE sync_id=$1`, syncID,
	).Scan(&r.SyncID, &r.KbID, &r.SourceType, &r.SourceConfig,
		&r.LastSyncedAt, &r.SyncStatus, &r.SyncError,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *StorageDao) GetExternalSyncByKB(ctx context.Context, kbID string) (*ExternalSyncRow, error) {
	syncID := "sync_" + kbID
	return s.GetExternalSync(ctx, syncID)
}

func (s *StorageDao) ListExternalSyncs(ctx context.Context, kbID string) ([]ExternalSyncRow, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT sync_id, kb_id, source_type, COALESCE(source_config::text,'{}'),
		        last_synced_at, sync_status, COALESCE(sync_error,''),
		        created_at, updated_at
		 FROM kb_external_sync WHERE kb_id=$1 ORDER BY created_at DESC`, kbID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ExternalSyncRow
	for rows.Next() {
		var r ExternalSyncRow
		if err := rows.Scan(&r.SyncID, &r.KbID, &r.SourceType, &r.SourceConfig,
			&r.LastSyncedAt, &r.SyncStatus, &r.SyncError,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *StorageDao) UpdateExternalSyncStatus(ctx context.Context, syncID, status, errMsg string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE kb_external_sync SET sync_status=$1, sync_error=$2,
		 last_synced_at=CASE WHEN $1='SYNCED' THEN NOW() ELSE last_synced_at END,
		 updated_at=NOW()
		 WHERE sync_id=$3`,
		status, errMsg, syncID,
	)
	return err
}

// BuildDSN builds a PostgreSQL connection string.
func BuildDSN(host, user, password, dbName string, port int) string {
	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     fmt.Sprintf("%s:%d", host, port),
		Path:     dbName,
		RawQuery: "sslmode=disable",
	}).String()
}
