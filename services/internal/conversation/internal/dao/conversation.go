package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Conversation struct {
	ConvId         int64
	Type           string
	GroupId        sql.NullInt64
	Uid            sql.NullInt64
	PeerUid        sql.NullInt64
	Name           sql.NullString
	Avatar         sql.NullString
	LastMsgId      sql.NullInt64
	LastMsgSnippet sql.NullString
	LastMsgTime    sql.NullTime
	LastMsgSender  sql.NullInt64
	CreatedAt      time.Time
}

type ConvMember struct {
	ConvId    int64
	Uid       int64
	Mute      bool
	Pinned    bool
	IsActive  bool
	UpdatedAt time.Time
	CreatedAt time.Time
}

type ConversationDao struct {
	db *pgxpool.Pool
}

func NewConversationDao(db *pgxpool.Pool) *ConversationDao {
	return &ConversationDao{db: db}
}

func (d *ConversationDao) GetConversationById(ctx context.Context, convId int64) (*Conversation, error) {
	row := d.db.QueryRow(ctx,
		`SELECT conv_id, type, group_id, uid, peer_uid, name, avatar, last_msg_id, last_msg_snippet, last_msg_time, last_msg_sender, created_at
		 FROM conversations WHERE conv_id = $1`, convId)
	c := &Conversation{}
	err := row.Scan(&c.ConvId, &c.Type, &c.GroupId, &c.Uid, &c.PeerUid, &c.Name, &c.Avatar,
		&c.LastMsgId, &c.LastMsgSnippet, &c.LastMsgTime, &c.LastMsgSender, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (d *ConversationDao) GetSingleConversationByPair(ctx context.Context, uid, peerUid int64) (*Conversation, error) {
	if uid > peerUid {
		uid, peerUid = peerUid, uid
	}
	row := d.db.QueryRow(ctx,
		`SELECT conv_id, type, group_id, uid, peer_uid, name, avatar, last_msg_id, last_msg_snippet, last_msg_time, last_msg_sender, created_at
		 FROM conversations WHERE type = 'SINGLE' AND uid = $1 AND peer_uid = $2`, uid, peerUid)
	c := &Conversation{}
	err := row.Scan(&c.ConvId, &c.Type, &c.GroupId, &c.Uid, &c.PeerUid, &c.Name, &c.Avatar,
		&c.LastMsgId, &c.LastMsgSnippet, &c.LastMsgTime, &c.LastMsgSender, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (d *ConversationDao) CreateSingleConversation(ctx context.Context, uid, peerUid int64) (int64, error) {
	if uid > peerUid {
		uid, peerUid = peerUid, uid
	}
	var convId int64
	err := d.db.QueryRow(ctx,
		`INSERT INTO conversations (type, uid, peer_uid, created_at)
		 VALUES ('SINGLE', $1, $2, NOW())
		 RETURNING conv_id`, uid, peerUid).Scan(&convId)
	return convId, err
}

func (d *ConversationDao) CreateGroupConversation(ctx context.Context, groupId int64, name, avatar string) (int64, error) {
	var convId int64
	err := d.db.QueryRow(ctx,
		`INSERT INTO conversations (type, group_id, name, avatar, created_at)
		 VALUES ('GROUP', $1, $2, $3, NOW())
		 RETURNING conv_id`, groupId, name, avatar).Scan(&convId)
	return convId, err
}

func (d *ConversationDao) DeleteConversation(ctx context.Context, convId int64) error {
	_, err := d.db.Exec(ctx, `DELETE FROM conversations WHERE conv_id = $1`, convId)
	return err
}

func (d *ConversationDao) UpdateLastMessage(ctx context.Context, convId, msgId int64, snippet string, sender int64) error {
	_, err := d.db.Exec(ctx,
		`UPDATE conversations SET last_msg_id = $1, last_msg_snippet = $2, last_msg_time = NOW(), last_msg_sender = $3 WHERE conv_id = $4`,
		msgId, snippet, sender, convId)
	return err
}

func (d *ConversationDao) ListConversationsByUid(ctx context.Context, uid int64) ([]*Conversation, error) {
	rows, err := d.db.Query(ctx,
		`SELECT c.conv_id, c.type, c.group_id, c.uid, c.peer_uid, c.name, c.avatar,
		        c.last_msg_id, c.last_msg_snippet, c.last_msg_time, c.last_msg_sender, c.created_at
		 FROM conversations c
		 JOIN conv_members cm ON c.conv_id = cm.conv_id
		 WHERE cm.uid = $1 AND cm.is_active = TRUE
		 ORDER BY c.last_msg_time DESC NULLS LAST, c.conv_id DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Conversation
	for rows.Next() {
		c := &Conversation{}
		err := rows.Scan(&c.ConvId, &c.Type, &c.GroupId, &c.Uid, &c.PeerUid, &c.Name, &c.Avatar,
			&c.LastMsgId, &c.LastMsgSnippet, &c.LastMsgTime, &c.LastMsgSender, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (d *ConversationDao) GetConvMember(ctx context.Context, convId, uid int64) (*ConvMember, error) {
	row := d.db.QueryRow(ctx,
		`SELECT conv_id, uid, mute, pinned, is_active, updated_at, created_at
		 FROM conv_members WHERE conv_id = $1 AND uid = $2`, convId, uid)
	m := &ConvMember{}
	err := row.Scan(&m.ConvId, &m.Uid, &m.Mute, &m.Pinned, &m.IsActive, &m.UpdatedAt, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (d *ConversationDao) UpsertConvMember(ctx context.Context, convId, uid int64, mute, pinned, isActive bool) error {
	_, err := d.db.Exec(ctx,
		`INSERT INTO conv_members (conv_id, uid, mute, pinned, is_active, updated_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		 ON CONFLICT (conv_id, uid) DO UPDATE SET mute = EXCLUDED.mute, pinned = EXCLUDED.pinned, is_active = EXCLUDED.is_active, updated_at = NOW()`,
		convId, uid, mute, pinned, isActive)
	return err
}

func (d *ConversationDao) SetConvMemberActive(ctx context.Context, convId, uid int64, isActive bool) error {
	_, err := d.db.Exec(ctx,
		`UPDATE conv_members SET is_active = $1, updated_at = NOW() WHERE conv_id = $2 AND uid = $3`,
		isActive, convId, uid)
	return err
}

func (d *ConversationDao) SetConvMemberMute(ctx context.Context, convId, uid int64, mute bool) error {
	_, err := d.db.Exec(ctx,
		`UPDATE conv_members SET mute = $1, updated_at = NOW() WHERE conv_id = $2 AND uid = $3`,
		mute, convId, uid)
	return err
}

func (d *ConversationDao) SetConvMemberPin(ctx context.Context, convId, uid int64, pinned bool) error {
	_, err := d.db.Exec(ctx,
		`UPDATE conv_members SET pinned = $1, updated_at = NOW() WHERE conv_id = $2 AND uid = $3`,
		pinned, convId, uid)
	return err
}

func (d *ConversationDao) ListConvMembers(ctx context.Context, convId int64) ([]int64, error) {
	rows, err := d.db.Query(ctx,
		`SELECT uid FROM conv_members WHERE conv_id = $1 AND is_active = TRUE`, convId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uids []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		uids = append(uids, uid)
	}
	return uids, rows.Err()
}

func (d *ConversationDao) BatchAddConvMembers(ctx context.Context, convId int64, uids []int64) error {
	batch := &pgx.Batch{}
	for _, uid := range uids {
		batch.Queue(
			`INSERT INTO conv_members (conv_id, uid, mute, pinned, is_active, updated_at, created_at)
			 VALUES ($1, $2, FALSE, FALSE, TRUE, NOW(), NOW())
			 ON CONFLICT (conv_id, uid) DO UPDATE SET is_active = TRUE, updated_at = NOW()`,
			convId, uid)
	}
	br := d.db.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < len(uids); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

func (d *ConversationDao) BatchRemoveConvMembers(ctx context.Context, convId int64, uids []int64) error {
	batch := &pgx.Batch{}
	for _, uid := range uids {
		batch.Queue(
			`UPDATE conv_members SET is_active = FALSE, updated_at = NOW() WHERE conv_id = $1 AND uid = $2`,
			convId, uid)
	}
	br := d.db.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < len(uids); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

func (d *ConversationDao) GetGroupConversationByGroupId(ctx context.Context, groupId int64) (*Conversation, error) {
	row := d.db.QueryRow(ctx,
		`SELECT conv_id, type, group_id, uid, peer_uid, name, avatar, last_msg_id, last_msg_snippet, last_msg_time, last_msg_sender, created_at
		 FROM conversations WHERE type = 'GROUP' AND group_id = $1`, groupId)
	c := &Conversation{}
	err := row.Scan(&c.ConvId, &c.Type, &c.GroupId, &c.Uid, &c.PeerUid, &c.Name, &c.Avatar,
		&c.LastMsgId, &c.LastMsgSnippet, &c.LastMsgTime, &c.LastMsgSender, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (d *ConversationDao) CountUnreadByUid(ctx context.Context, uid int64) (int64, error) {
	var count int64
	err := d.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(unread_count), 0) FROM user_unread WHERE uid = $1`, uid).Scan(&count)
	return count, err
}

func (d *ConversationDao) ClearUnread(ctx context.Context, convId, uid int64) error {
	_, err := d.db.Exec(ctx,
		`INSERT INTO user_unread (uid, conv_id, unread_count, updated_at)
		 VALUES ($1, $2, 0, NOW())
		 ON CONFLICT (uid, conv_id) DO UPDATE SET unread_count = 0, updated_at = NOW()`,
		uid, convId)
	return err
}

func (d *ConversationDao) IncrementUnread(ctx context.Context, convId int64, uids []int64) error {
	batch := &pgx.Batch{}
	for _, uid := range uids {
		batch.Queue(
			`INSERT INTO user_unread (uid, conv_id, unread_count, updated_at)
			 VALUES ($1, $2, 1, NOW())
			 ON CONFLICT (uid, conv_id) DO UPDATE SET unread_count = user_unread.unread_count + 1, updated_at = NOW()`,
			uid, convId)
	}
	br := d.db.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < len(uids); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return br.Close()
}

func (d *ConversationDao) GetUnreadCount(ctx context.Context, convId, uid int64) (int64, error) {
	var count int64
	err := d.db.QueryRow(ctx,
		`SELECT COALESCE(unread_count, 0) FROM user_unread WHERE uid = $1 AND conv_id = $2`, uid, convId).Scan(&count)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return count, err
}

func (d *ConversationDao) DeleteConvMember(ctx context.Context, convId, uid int64) error {
	_, err := d.db.Exec(ctx,
		`DELETE FROM conv_members WHERE conv_id = $1 AND uid = $2`, convId, uid)
	return err
}
