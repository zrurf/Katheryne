package logic

import (
	"crypto/sha256"
	"fmt"
	"time"

	"mem/internal/dao"

	"mem/mem/mem"

	"github.com/google/uuid"
)

// generateMemoryID creates a unique memory ID.
func generateMemoryID() string {
	return "mem_" + uuid.New().String()[:8]
}

// generateEdgeID creates a unique graph edge ID.
func generateEdgeID(source, relType, target string) string {
	h := sha256.Sum256([]byte(source + relType + target))
	return fmt.Sprintf("edge_%x", h[:8])
}

// toProto converts a DAO MemoryRow to a proto MemoryItem.
func toProto(row *dao.MemoryRow) *mem.MemoryItem {
	item := &mem.MemoryItem{
		MemoryId:       row.MemoryID,
		TenantId:       row.TenantID,
		TenantType:     row.TenantType,
		Type:           mem.MemoryType(row.MemoryType),
		Content:        row.Content,
		Importance:     row.Importance,
		CreatedAt:      row.CreatedAt.Unix(),
		UpdatedAt:      row.UpdatedAt.Unix(),
		LastAccessedAt: row.LastAccessedAt.Unix(),
		AccessCount:    row.AccessCount,
		Entities:       row.Entities,
		Metadata:       row.Metadata,
	}
	if row.ExpiresAt != nil {
		item.ExpiresAt = row.ExpiresAt.Unix()
	}
	if item.Entities == nil {
		item.Entities = []string{}
	}
	if item.Metadata == nil {
		item.Metadata = make(map[string]string)
	}
	return item
}

// computeExpiry converts TTL seconds to a time pointer.
func computeExpiry(ttlSeconds int64) *time.Time {
	if ttlSeconds <= 0 {
		return nil
	}
	t := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	return &t
}

// ptrFloat64 returns a pointer to a float64 value.
func ptrFloat64(v float64) *float64 {
	return &v
}