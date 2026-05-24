package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	Id          int64
	ConvId      int64
	Sender      int64
	Receiver    int64
	Type        string
	Content     string
	ContentType string
	QuoteMsgId  sql.NullInt64
	Recalled    bool
	RecallTime  sql.NullTime
	Edited      bool
	EditedAt    sql.NullTime
	EditCount   int32
	Extra       sql.NullString
	CreatedAt   time.Time
}

type ReadInterval struct {
	Id         int64
	ConvId     int64
	Uid        int64
	StartMsgId int64
	EndMsgId   int64
	CreatedAt  time.Time
}

type MessageDao struct {
	db *pgxpool.Pool
}

func NewMessageDao(db *pgxpool.Pool) *MessageDao {
	return &MessageDao{db: db}
}

func (d *MessageDao) InsertMessage(ctx context.Context, convId, sender, receiver int64, msgType, content, contentType string, quoteMsgId int64, extra string) (*Message, error) {
	var quote sql.NullInt64
	if quoteMsgId > 0 {
		quote = sql.NullInt64{Int64: quoteMsgId, Valid: true}
	}
	var extraNull sql.NullString
	if extra != "" {
		extraNull = sql.NullString{String: extra, Valid: true}
	}

	m := &Message{
		ConvId:      convId,
		Sender:      sender,
		Receiver:    receiver,
		Type:        msgType,
		Content:     content,
		ContentType: contentType,
		QuoteMsgId:  quote,
		Extra:       extraNull,
	}

	err := d.db.QueryRow(ctx,
		`INSERT INTO message (conv_id, sender, receiver, type, content, content_type, quote_msg_id, extra, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		 RETURNING id, created_at`,
		m.ConvId, m.Sender, m.Receiver, m.Type, m.Content, m.ContentType, m.QuoteMsgId, m.Extra,
	).Scan(&m.Id, &m.CreatedAt)
	return m, err
}

func (d *MessageDao) GetMessageById(ctx context.Context, msgId int64) (*Message, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id, conv_id, sender, receiver, type, content, content_type, quote_msg_id, recalled, recall_time, edited, edited_at, edit_count, extra, created_at
		 FROM message WHERE id = $1`, msgId)
	m := &Message{}
	err := row.Scan(&m.Id, &m.ConvId, &m.Sender, &m.Receiver, &m.Type, &m.Content, &m.ContentType,
		&m.QuoteMsgId, &m.Recalled, &m.RecallTime, &m.Edited, &m.EditedAt, &m.EditCount, &m.Extra, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (d *MessageDao) RecallMessage(ctx context.Context, msgId int64) error {
	_, err := d.db.Exec(ctx,
		`UPDATE message SET recalled = TRUE, recall_time = NOW() WHERE id = $1`, msgId)
	return err
}

func (d *MessageDao) EditMessage(ctx context.Context, msgId int64, content string, extra string) error {
	_, err := d.db.Exec(ctx,
		`UPDATE message SET content = $1, extra = $2::jsonb, edited = TRUE, edited_at = NOW(), edit_count = edit_count + 1 WHERE id = $3`,
		content, extra, msgId)
	return err
}

func (d *MessageDao) GetMessagesBefore(ctx context.Context, convId int64, cursor int64, limit int32) ([]*Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := d.db.Query(ctx,
		`SELECT id, conv_id, sender, receiver, type, content, content_type, quote_msg_id, recalled, recall_time, edited, edited_at, edit_count, extra, created_at
		 FROM message WHERE conv_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`,
		convId, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (d *MessageDao) GetMessagesAfter(ctx context.Context, convId int64, cursor int64, limit int32) ([]*Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := d.db.Query(ctx,
		`SELECT id, conv_id, sender, receiver, type, content, content_type, quote_msg_id, recalled, recall_time, edited, edited_at, edit_count, extra, created_at
		 FROM message WHERE conv_id = $1 AND id > $2 ORDER BY id ASC LIMIT $3`,
		convId, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (d *MessageDao) GetLatestMessages(ctx context.Context, convId int64, limit int32) ([]*Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := d.db.Query(ctx,
		`SELECT id, conv_id, sender, receiver, type, content, content_type, quote_msg_id, recalled, recall_time, edited, edited_at, edit_count, extra, created_at
		 FROM message WHERE conv_id = $1 ORDER BY id DESC LIMIT $2`,
		convId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (d *MessageDao) SearchMessages(ctx context.Context, keyword string, convId, sender int64, startTime, endTime int64, page, size int32) ([]*Message, int64, error) {
	if size <= 0 || size > 100 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * size

	where := "WHERE recalled = FALSE"
	args := []interface{}{}
	argIdx := 1

	if keyword != "" {
		where += fmt.Sprintf(" AND to_tsvector('simple', content) @@ plainto_tsquery('simple', $%d)", argIdx)
		args = append(args, keyword)
		argIdx++
	}
	if convId > 0 {
		where += fmt.Sprintf(" AND conv_id = $%d", argIdx)
		args = append(args, convId)
		argIdx++
	}
	if sender > 0 {
		where += fmt.Sprintf(" AND sender = $%d", argIdx)
		args = append(args, sender)
		argIdx++
	}
	if startTime > 0 {
		where += fmt.Sprintf(" AND created_at >= TO_TIMESTAMP($%d::double precision / 1000)", argIdx)
		args = append(args, startTime)
		argIdx++
	}
	if endTime > 0 {
		where += fmt.Sprintf(" AND created_at <= TO_TIMESTAMP($%d::double precision / 1000)", argIdx)
		args = append(args, endTime)
		argIdx++
	}

	countSQL := "SELECT COUNT(*) FROM message " + where
	var total int64
	err := d.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(
		`SELECT id, conv_id, sender, receiver, type, content, content_type, quote_msg_id, recalled, recall_time, edited, edited_at, edit_count, extra, created_at
		 FROM message %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := d.db.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list, err := scanMessages(rows)
	return list, total, err
}

func (d *MessageDao) SubmitReadInterval(ctx context.Context, convId, uid, startMsgId, endMsgId int64) error {
	if startMsgId <= 0 {
		startMsgId = endMsgId
	}
	if endMsgId <= 0 {
		endMsgId = startMsgId
	}
	if startMsgId > endMsgId {
		startMsgId, endMsgId = endMsgId, startMsgId
	}

	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var minStart, maxEnd int64
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MIN(start_msg_id), $3) as min_start, COALESCE(MAX(end_msg_id), $4) as max_end
		 FROM msg_read_intervals
		 WHERE conv_id = $1 AND uid = $2
		   AND start_msg_id <= $4 AND end_msg_id >= $3`,
		convId, uid, startMsgId, endMsgId,
	).Scan(&minStart, &maxEnd)
	if err != nil {
		return err
	}

	if minStart < startMsgId {
		startMsgId = minStart
	}
	if maxEnd > endMsgId {
		endMsgId = maxEnd
	}

	_, err = tx.Exec(ctx,
		`DELETE FROM msg_read_intervals
		 WHERE conv_id = $1 AND uid = $2
		   AND start_msg_id >= $3 AND end_msg_id <= $4`,
		convId, uid, startMsgId, endMsgId,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO msg_read_intervals (conv_id, uid, start_msg_id, end_msg_id, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (conv_id, uid, start_msg_id) DO UPDATE SET end_msg_id = GREATEST(msg_read_intervals.end_msg_id, EXCLUDED.end_msg_id), created_at = NOW()`,
		convId, uid, startMsgId, endMsgId,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (d *MessageDao) GetReadMembersByMsgId(ctx context.Context, convId, msgId int64) ([]*ReadInterval, error) {
	rows, err := d.db.Query(ctx,
		`SELECT id, conv_id, uid, start_msg_id, end_msg_id, created_at
		 FROM msg_read_intervals WHERE conv_id = $1 AND start_msg_id <= $2 AND end_msg_id >= $2`,
		convId, msgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ReadInterval
	for rows.Next() {
		r := &ReadInterval{}
		err := rows.Scan(&r.Id, &r.ConvId, &r.Uid, &r.StartMsgId, &r.EndMsgId, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (d *MessageDao) GetReadIntervals(ctx context.Context, convId, uid int64) ([]*ReadInterval, error) {
	rows, err := d.db.Query(ctx,
		`SELECT id, conv_id, uid, start_msg_id, end_msg_id, created_at
		 FROM msg_read_intervals WHERE conv_id = $1 AND uid = $2
		 ORDER BY start_msg_id ASC`,
		convId, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ReadInterval
	for rows.Next() {
		r := &ReadInterval{}
		err := rows.Scan(&r.Id, &r.ConvId, &r.Uid, &r.StartMsgId, &r.EndMsgId, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (d *MessageDao) GetUnreadCount(ctx context.Context, convId, uid int64) (int64, error) {
	var count int64
	err := d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM message m
		 WHERE m.conv_id = $1 AND m.id > COALESCE((SELECT MAX(end_msg_id) FROM msg_read_intervals WHERE conv_id = $1 AND uid = $2), 0)`,
		convId, uid).Scan(&count)
	return count, err
}

func (d *MessageDao) BatchGetUnreadCount(ctx context.Context, uid int64, convIds []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(convIds))
	for _, convId := range convIds {
		count, err := d.GetUnreadCount(ctx, convId, uid)
		if err != nil {
			return nil, err
		}
		result[convId] = count
	}
	return result, nil
}

func (d *MessageDao) SyncOfflineMessages(ctx context.Context, uid int64, lastSyncMsgId int64, limit int32, convIds []int64) ([]*Message, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if len(convIds) == 0 {
		return nil, nil
	}
	rows, err := d.db.Query(ctx,
		`SELECT id, conv_id, sender, receiver, type, content, content_type, quote_msg_id, recalled, recall_time, edited, edited_at, edit_count, extra, created_at
		 FROM message
		 WHERE id > $1 AND conv_id = ANY($2)
		 ORDER BY id ASC LIMIT $3`,
		lastSyncMsgId, convIds, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (d *MessageDao) GetLastMessage(ctx context.Context, convId int64) (*Message, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id, conv_id, sender, receiver, type, content, content_type, quote_msg_id, recalled, recall_time, edited, edited_at, edit_count, extra, created_at
		 FROM message WHERE conv_id = $1 ORDER BY id DESC LIMIT 1`, convId)
	m := &Message{}
	err := row.Scan(&m.Id, &m.ConvId, &m.Sender, &m.Receiver, &m.Type, &m.Content, &m.ContentType,
		&m.QuoteMsgId, &m.Recalled, &m.RecallTime, &m.Edited, &m.EditedAt, &m.EditCount, &m.Extra, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func scanMessages(rows pgx.Rows) ([]*Message, error) {
	var list []*Message
	for rows.Next() {
		m := &Message{}
		err := rows.Scan(&m.Id, &m.ConvId, &m.Sender, &m.Receiver, &m.Type, &m.Content, &m.ContentType,
			&m.QuoteMsgId, &m.Recalled, &m.RecallTime, &m.Edited, &m.EditedAt, &m.EditCount, &m.Extra, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}
