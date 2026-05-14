package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ========== 好友 ==========

type Friendship struct {
	Uid       int64
	PeerUid   int64
	Remark    sql.NullString
	GroupName sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FriendRequest struct {
	Id        int64
	Uid       int64
	ToUid     int64
	Message   sql.NullString
	Status    string
	CreatedAt time.Time
}

// ========== 黑名单 ==========

type Blacklist struct {
	Uid       int64
	PeerUid   int64
	CreatedAt time.Time
}

// ========== 群组 ==========

type Group struct {
	Id          int64
	GroupId     int64
	Name        string
	Avatar      sql.NullString
	Owner       int64
	MemberCount int32
	Status      string
	VerifyMode  string
	CreatedAt   time.Time
}

type GroupMember struct {
	GroupId   int64
	Uid       int64
	Role      string
	Nick      sql.NullString
	JoinTime  time.Time
	Inviter   sql.NullInt64
	MuteUntil sql.NullTime
}

type GroupAnnouncement struct {
	Id        int64
	GroupId   int64
	Uid       int64
	Content   string
	Pinned    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ========== 用户 ==========

type User struct {
	Uid       int64
	Name      string
	Phone     string
	Avatar    sql.NullString
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastLogin time.Time
}

// ========== DAO ==========

type SocialDao struct {
	db *pgxpool.Pool
}

func NewSocialDao(db *pgxpool.Pool) *SocialDao {
	return &SocialDao{db: db}
}

// --- 好友请求 ---

func (d *SocialDao) InsertFriendRequest(ctx context.Context, uid, toUid int64, msg string) (*FriendRequest, error) {
	var msgNull sql.NullString
	if msg != "" {
		msgNull = sql.NullString{String: msg, Valid: true}
	}
	req := &FriendRequest{Uid: uid, ToUid: toUid, Message: msgNull, Status: "pending"}
	err := d.db.QueryRow(ctx,
		`INSERT INTO friend_requests (uid, to_uid, message, status, created_at)
		 VALUES ($1, $2, $3, 'pending', NOW())
		 ON CONFLICT (uid, to_uid) DO UPDATE SET message = EXCLUDED.message, status = 'pending', created_at = NOW()
		 RETURNING id, created_at`,
		uid, toUid, msgNull,
	).Scan(&req.Id, &req.CreatedAt)
	return req, err
}

func (d *SocialDao) GetFriendRequestById(ctx context.Context, reqId int64) (*FriendRequest, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id, uid, to_uid, message, status, created_at FROM friend_requests WHERE id = $1`, reqId)
	r := &FriendRequest{}
	err := row.Scan(&r.Id, &r.Uid, &r.ToUid, &r.Message, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (d *SocialDao) UpdateFriendRequestStatus(ctx context.Context, reqId int64, status string) error {
	_, err := d.db.Exec(ctx,
		`UPDATE friend_requests SET status = $1 WHERE id = $2`, status, reqId)
	return err
}

func (d *SocialDao) ListFriendRequests(ctx context.Context, uid int64, reqType string, page, size int32) ([]*FriendRequest, int64, error) {
	if size <= 0 || size > 100 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * size

	var where string
	var args []interface{}
	if reqType == "sent" {
		where = "WHERE uid = $1"
		args = append(args, uid)
	} else {
		where = "WHERE to_uid = $1"
		args = append(args, uid)
	}

	var total int64
	err := d.db.QueryRow(ctx, `SELECT COUNT(*) FROM friend_requests `+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := d.db.Query(ctx,
		`SELECT id, uid, to_uid, message, status, created_at FROM friend_requests `+where+` ORDER BY id DESC LIMIT $2 OFFSET $3`,
		args[0], size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*FriendRequest
	for rows.Next() {
		r := &FriendRequest{}
		err := rows.Scan(&r.Id, &r.Uid, &r.ToUid, &r.Message, &r.Status, &r.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, rows.Err()
}

// --- 好友关系 ---

func (d *SocialDao) AddFriendship(ctx context.Context, uid, peerUid int64, remark, groupName string) error {
	if uid > peerUid {
		uid, peerUid = peerUid, uid
	}

	var r, g sql.NullString
	if remark != "" {
		r = sql.NullString{String: remark, Valid: true}
	}
	if groupName != "" {
		g = sql.NullString{String: groupName, Valid: true}
	}

	_, err := d.db.Exec(ctx,
		`INSERT INTO friendship (uid, peer_uid, remark, group_name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 ON CONFLICT (uid, peer_uid) DO UPDATE SET remark = EXCLUDED.remark, group_name = EXCLUDED.group_name, updated_at = NOW()`,
		uid, peerUid, r, g)
	return err
}

func (d *SocialDao) DeleteFriendship(ctx context.Context, uid, peerUid int64) error {
	if uid > peerUid {
		uid, peerUid = peerUid, uid
	}
	_, err := d.db.Exec(ctx,
		`DELETE FROM friendship WHERE uid = $1 AND peer_uid = $2`, uid, peerUid)
	return err
}

func (d *SocialDao) UpdateFriendRemark(ctx context.Context, uid, peerUid int64, remark, groupName string) error {
	if uid > peerUid {
		uid, peerUid = peerUid, uid
	}

	var setParts []string
	var args []interface{}
	argIdx := 1

	if remark != "" {
		setParts = append(setParts, "remark = $"+itoa(argIdx))
		args = append(args, remark)
		argIdx++
	}
	if groupName != "" {
		setParts = append(setParts, "group_name = $"+itoa(argIdx))
		args = append(args, groupName)
		argIdx++
	}
	if len(setParts) == 0 {
		return nil
	}

	args = append(args, uid, peerUid)
	_, err := d.db.Exec(ctx,
		`UPDATE friendship SET `+join(setParts, ", ")+`, updated_at = NOW() WHERE uid = $`+itoa(argIdx)+` AND peer_uid = $`+itoa(argIdx+1),
		args...)
	return err
}

func (d *SocialDao) ListFriends(ctx context.Context, uid int64, group string) ([]*Friendship, error) {
	var rows pgx.Rows
	var err error
	if group != "" {
		rows, err = d.db.Query(ctx,
			`SELECT uid, peer_uid, remark, group_name, created_at, updated_at FROM friendship
			 WHERE (uid = $1 OR peer_uid = $1) AND group_name = $2`, uid, group)
	} else {
		rows, err = d.db.Query(ctx,
			`SELECT uid, peer_uid, remark, group_name, created_at, updated_at FROM friendship
			 WHERE uid = $1 OR peer_uid = $1`, uid)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Friendship
	for rows.Next() {
		f := &Friendship{}
		err := rows.Scan(&f.Uid, &f.PeerUid, &f.Remark, &f.GroupName, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if f.PeerUid == uid {
			f.Uid, f.PeerUid = f.PeerUid, f.Uid
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

func (d *SocialDao) IsFriend(ctx context.Context, uid, peerUid int64) (bool, error) {
	if uid > peerUid {
		uid, peerUid = peerUid, uid
	}
	var count int64
	err := d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM friendship WHERE uid = $1 AND peer_uid = $2`, uid, peerUid).Scan(&count)
	return count > 0, err
}

// --- 黑名单 ---

func (d *SocialDao) AddBlacklist(ctx context.Context, uid, peerUid int64) error {
	_, err := d.db.Exec(ctx,
		`INSERT INTO blacklist (uid, peer_uid, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`,
		uid, peerUid)
	return err
}

func (d *SocialDao) RemoveBlacklist(ctx context.Context, uid, peerUid int64) error {
	_, err := d.db.Exec(ctx,
		`DELETE FROM blacklist WHERE uid = $1 AND peer_uid = $2`, uid, peerUid)
	return err
}

func (d *SocialDao) IsBlacklisted(ctx context.Context, uid, peerUid int64) (bool, error) {
	var count int64
	err := d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM blacklist WHERE uid = $1 AND peer_uid = $2`, uid, peerUid).Scan(&count)
	return count > 0, err
}

func (d *SocialDao) GetBlacklist(ctx context.Context, uid int64) ([]*Blacklist, error) {
	rows, err := d.db.Query(ctx,
		`SELECT uid, peer_uid, created_at FROM blacklist WHERE uid = $1`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Blacklist
	for rows.Next() {
		b := &Blacklist{}
		err := rows.Scan(&b.Uid, &b.PeerUid, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

// --- 群组 ---

func (d *SocialDao) InsertGroup(ctx context.Context, groupId, owner int64, name, avatar, verifyMode string) (*Group, error) {
	var avatarNull sql.NullString
	if avatar != "" {
		avatarNull = sql.NullString{String: avatar, Valid: true}
	}
	g := &Group{
		GroupId:    groupId,
		Name:       name,
		Avatar:     avatarNull,
		Owner:      owner,
		Status:     "ACTIVE",
		VerifyMode: verifyMode,
	}
	err := d.db.QueryRow(ctx,
		`INSERT INTO groups (group_id, name, avatar, owner, member_count, status, verify_mode, created_at)
		 VALUES ($1, $2, $3, $4, 1, 'ACTIVE', $5, NOW())
		 RETURNING id, created_at`,
		g.GroupId, g.Name, g.Avatar, g.Owner, g.VerifyMode,
	).Scan(&g.Id, &g.CreatedAt)
	return g, err
}

func (d *SocialDao) GetGroupById(ctx context.Context, groupId int64) (*Group, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id, group_id, name, avatar, owner, member_count, status, verify_mode, created_at FROM groups WHERE group_id = $1`, groupId)
	g := &Group{}
	err := row.Scan(&g.Id, &g.GroupId, &g.Name, &g.Avatar, &g.Owner, &g.MemberCount, &g.Status, &g.VerifyMode, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (d *SocialDao) UpdateGroup(ctx context.Context, groupId int64, name, avatar, verifyMode string) error {
	var setParts []string
	var args []interface{}
	argIdx := 1

	if name != "" {
		setParts = append(setParts, "name = $"+itoa(argIdx))
		args = append(args, name)
		argIdx++
	}
	if avatar != "" {
		setParts = append(setParts, "avatar = $"+itoa(argIdx))
		args = append(args, avatar)
		argIdx++
	}
	if verifyMode != "" {
		setParts = append(setParts, "verify_mode = $"+itoa(argIdx))
		args = append(args, verifyMode)
		argIdx++
	}
	if len(setParts) == 0 {
		return nil
	}

	args = append(args, groupId)
	_, err := d.db.Exec(ctx,
		`UPDATE groups SET `+join(setParts, ", ")+` WHERE group_id = $`+itoa(argIdx),
		args...)
	return err
}

func (d *SocialDao) UpdateGroupOwner(ctx context.Context, groupId, newOwner int64) error {
	_, err := d.db.Exec(ctx,
		`UPDATE groups SET owner = $1 WHERE group_id = $2`, newOwner, groupId)
	return err
}

func (d *SocialDao) DeleteGroup(ctx context.Context, groupId int64) error {
	_, err := d.db.Exec(ctx,
		`DELETE FROM groups WHERE group_id = $1`, groupId)
	return err
}

// --- 群成员 ---

func (d *SocialDao) AddGroupMember(ctx context.Context, groupId, uid int64, role, nick string, inviter int64) error {
	var nickNull sql.NullString
	if nick != "" {
		nickNull = sql.NullString{String: nick, Valid: true}
	}
	var inviterNull sql.NullInt64
	if inviter > 0 {
		inviterNull = sql.NullInt64{Int64: inviter, Valid: true}
	}
	_, err := d.db.Exec(ctx,
		`INSERT INTO group_members (group_id, uid, role, nick, join_time, inviter, mute_until)
		 VALUES ($1, $2, $3, $4, NOW(), $5, NULL)
		 ON CONFLICT (group_id, uid) DO UPDATE SET role = EXCLUDED.role, nick = EXCLUDED.nick`,
		groupId, uid, role, nickNull, inviterNull)
	return err
}

func (d *SocialDao) RemoveGroupMember(ctx context.Context, groupId, uid int64) error {
	_, err := d.db.Exec(ctx,
		`DELETE FROM group_members WHERE group_id = $1 AND uid = $2`, groupId, uid)
	return err
}

func (d *SocialDao) GetGroupMember(ctx context.Context, groupId, uid int64) (*GroupMember, error) {
	row := d.db.QueryRow(ctx,
		`SELECT group_id, uid, role, nick, join_time, inviter, mute_until FROM group_members WHERE group_id = $1 AND uid = $2`,
		groupId, uid)
	m := &GroupMember{}
	err := row.Scan(&m.GroupId, &m.Uid, &m.Role, &m.Nick, &m.JoinTime, &m.Inviter, &m.MuteUntil)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (d *SocialDao) ListGroupMembers(ctx context.Context, groupId int64, role string) ([]*GroupMember, error) {
	var rows pgx.Rows
	var err error
	if role != "" {
		rows, err = d.db.Query(ctx,
			`SELECT group_id, uid, role, nick, join_time, inviter, mute_until FROM group_members WHERE group_id = $1 AND role = $2 ORDER BY join_time`,
			groupId, role)
	} else {
		rows, err = d.db.Query(ctx,
			`SELECT group_id, uid, role, nick, join_time, inviter, mute_until FROM group_members WHERE group_id = $1 ORDER BY join_time`,
			groupId)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*GroupMember
	for rows.Next() {
		m := &GroupMember{}
		err := rows.Scan(&m.GroupId, &m.Uid, &m.Role, &m.Nick, &m.JoinTime, &m.Inviter, &m.MuteUntil)
		if err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (d *SocialDao) UpdateGroupMemberRole(ctx context.Context, groupId, uid int64, role string) error {
	_, err := d.db.Exec(ctx,
		`UPDATE group_members SET role = $1 WHERE group_id = $2 AND uid = $3`, role, groupId, uid)
	return err
}

func (d *SocialDao) UpdateMemberMute(ctx context.Context, groupId, uid int64, muteUntil *time.Time) error {
	var t sql.NullTime
	if muteUntil != nil {
		t = sql.NullTime{Time: *muteUntil, Valid: true}
	}
	_, err := d.db.Exec(ctx,
		`UPDATE group_members SET mute_until = $1 WHERE group_id = $2 AND uid = $3`, t, groupId, uid)
	return err
}

func (d *SocialDao) IncrGroupMemberCount(ctx context.Context, groupId int64, delta int32) error {
	_, err := d.db.Exec(ctx,
		`UPDATE groups SET member_count = member_count + $1 WHERE group_id = $2`, delta, groupId)
	return err
}

// --- 群公告 ---

func (d *SocialDao) InsertAnnouncement(ctx context.Context, groupId, uid int64, content string) (*GroupAnnouncement, error) {
	a := &GroupAnnouncement{GroupId: groupId, Uid: uid, Content: content}
	err := d.db.QueryRow(ctx,
		`INSERT INTO group_announcement (group_id, uid, content, pinned, created_at, updated_at)
		 VALUES ($1, $2, $3, FALSE, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		groupId, uid, content,
	).Scan(&a.Id, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (d *SocialDao) ListAnnouncements(ctx context.Context, groupId int64, page, size int32) ([]*GroupAnnouncement, int64, error) {
	if size <= 0 || size > 100 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * size

	var total int64
	err := d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM group_announcement WHERE group_id = $1`, groupId).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := d.db.Query(ctx,
		`SELECT id, group_id, uid, content, pinned, created_at, updated_at FROM group_announcement
		 WHERE group_id = $1 ORDER BY pinned DESC, id DESC LIMIT $2 OFFSET $3`,
		groupId, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*GroupAnnouncement
	for rows.Next() {
		a := &GroupAnnouncement{}
		err := rows.Scan(&a.Id, &a.GroupId, &a.Uid, &a.Content, &a.Pinned, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}
	return list, total, rows.Err()
}

// --- 群加入申请 ---

type GroupJoinRequest struct {
	Id        int64
	GroupId   int64
	Uid       int64
	Message   sql.NullString
	Status    string
	CreatedAt time.Time
}

func (d *SocialDao) InsertGroupJoinRequest(ctx context.Context, groupId, uid int64, msg string) (*GroupJoinRequest, error) {
	var msgNull sql.NullString
	if msg != "" {
		msgNull = sql.NullString{String: msg, Valid: true}
	}
	req := &GroupJoinRequest{GroupId: groupId, Uid: uid, Message: msgNull, Status: "pending"}
	err := d.db.QueryRow(ctx,
		`INSERT INTO group_join_requests (group_id, uid, message, status, created_at)
		 VALUES ($1, $2, $3, 'pending', NOW())
		 ON CONFLICT (group_id, uid) DO UPDATE SET message = EXCLUDED.message, status = 'pending', created_at = NOW()
		 RETURNING id, created_at`,
		groupId, uid, msgNull,
	).Scan(&req.Id, &req.CreatedAt)
	return req, err
}

func (d *SocialDao) GetGroupJoinRequestById(ctx context.Context, reqId int64) (*GroupJoinRequest, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id, group_id, uid, message, status, created_at FROM group_join_requests WHERE id = $1`, reqId)
	r := &GroupJoinRequest{}
	err := row.Scan(&r.Id, &r.GroupId, &r.Uid, &r.Message, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (d *SocialDao) UpdateGroupJoinRequestStatus(ctx context.Context, reqId int64, status string) error {
	_, err := d.db.Exec(ctx,
		`UPDATE group_join_requests SET status = $1 WHERE id = $2`, status, reqId)
	return err
}

func (d *SocialDao) ListGroupJoinRequests(ctx context.Context, groupId int64, status string, page, size int32) ([]*GroupJoinRequest, int64, error) {
	if size <= 0 || size > 100 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * size

	where := "WHERE group_id = $1"
	args := []interface{}{groupId}
	if status != "" {
		where += " AND status = $2"
		args = append(args, status)
	}

	var total int64
	err := d.db.QueryRow(ctx, `SELECT COUNT(*) FROM group_join_requests `+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, group_id, uid, message, status, created_at FROM group_join_requests ` + where + ` ORDER BY id DESC LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	args = append(args, size, offset)

	rows, err := d.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*GroupJoinRequest
	for rows.Next() {
		r := &GroupJoinRequest{}
		err := rows.Scan(&r.Id, &r.GroupId, &r.Uid, &r.Message, &r.Status, &r.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, rows.Err()
}

// --- 群邀请 ---

type GroupInvite struct {
	Id        int64
	GroupId   int64
	Inviter   int64
	Invitee   int64
	Message   sql.NullString
	Status    string
	CreatedAt time.Time
}

func (d *SocialDao) InsertGroupInvite(ctx context.Context, groupId, inviter, invitee int64, msg string) (*GroupInvite, error) {
	var msgNull sql.NullString
	if msg != "" {
		msgNull = sql.NullString{String: msg, Valid: true}
	}
	inv := &GroupInvite{GroupId: groupId, Inviter: inviter, Invitee: invitee, Message: msgNull, Status: "pending"}
	err := d.db.QueryRow(ctx,
		`INSERT INTO group_invites (group_id, inviter, invitee, message, status, created_at)
		 VALUES ($1, $2, $3, $4, 'pending', NOW())
		 ON CONFLICT (group_id, invitee) DO UPDATE SET inviter = EXCLUDED.inviter, message = EXCLUDED.message, status = 'pending', created_at = NOW()
		 RETURNING id, created_at`,
		groupId, inviter, invitee, msgNull,
	).Scan(&inv.Id, &inv.CreatedAt)
	return inv, err
}

func (d *SocialDao) GetGroupInviteById(ctx context.Context, inviteId int64) (*GroupInvite, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id, group_id, inviter, invitee, message, status, created_at FROM group_invites WHERE id = $1`, inviteId)
	inv := &GroupInvite{}
	err := row.Scan(&inv.Id, &inv.GroupId, &inv.Inviter, &inv.Invitee, &inv.Message, &inv.Status, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (d *SocialDao) UpdateGroupInviteStatus(ctx context.Context, inviteId int64, status string) error {
	_, err := d.db.Exec(ctx,
		`UPDATE group_invites SET status = $1 WHERE id = $2`, status, inviteId)
	return err
}

func (d *SocialDao) ListGroupInvitesByInvitee(ctx context.Context, invitee int64) ([]*GroupInvite, error) {
	rows, err := d.db.Query(ctx,
		`SELECT id, group_id, inviter, invitee, message, status, created_at FROM group_invites WHERE invitee = $1 AND status = 'pending' ORDER BY id DESC`,
		invitee)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*GroupInvite
	for rows.Next() {
		inv := &GroupInvite{}
		err := rows.Scan(&inv.Id, &inv.GroupId, &inv.Inviter, &inv.Invitee, &inv.Message, &inv.Status, &inv.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

// --- 会话（用于创建群聊时联动） ---

func (d *SocialDao) InsertConversation(ctx context.Context, convType string, groupId int64, name, avatar string) (int64, error) {
	var id int64
	err := d.db.QueryRow(ctx,
		`INSERT INTO conversations (type, group_id, name, avatar, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING conv_id`,
		convType, groupId, name, avatar).Scan(&id)
	return id, err
}

func (d *SocialDao) AddConvMember(ctx context.Context, convId, uid int64) error {
	_, err := d.db.Exec(ctx,
		`INSERT INTO conv_members (conv_id, uid, is_active, created_at, updated_at) VALUES ($1, $2, TRUE, NOW(), NOW()) ON CONFLICT DO NOTHING`,
		convId, uid)
	return err
}

func (d *SocialDao) GetConversationByGroupId(ctx context.Context, groupId int64) (int64, error) {
	var convId int64
	err := d.db.QueryRow(ctx,
		`SELECT conv_id FROM conversations WHERE group_id = $1 AND type = 'GROUP'`, groupId).Scan(&convId)
	return convId, err
}

// --- 用户群组 ---

func (d *SocialDao) ListUserGroups(ctx context.Context, uid int64) ([]*Group, error) {
	rows, err := d.db.Query(ctx,
		`SELECT g.id, g.group_id, g.name, g.avatar, g.owner, g.member_count, g.status, g.verify_mode, g.created_at
		 FROM groups g
		 JOIN group_members gm ON g.group_id = gm.group_id
		 WHERE gm.uid = $1 AND g.status = 'ACTIVE'
		 ORDER BY g.created_at DESC`,
		uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Group
	for rows.Next() {
		g := &Group{}
		err := rows.Scan(&g.Id, &g.GroupId, &g.Name, &g.Avatar, &g.Owner, &g.MemberCount, &g.Status, &g.VerifyMode, &g.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}

// --- 群公告更新 ---

func (d *SocialDao) UpdateGroupAnnouncement(ctx context.Context, groupId int64, content string) error {
	_, err := d.db.Exec(ctx,
		`INSERT INTO group_announcement (group_id, uid, content, pinned, created_at, updated_at)
		 VALUES ($1, 0, $2, TRUE, NOW(), NOW())
		 ON CONFLICT (group_id) DO UPDATE SET content = EXCLUDED.content, updated_at = NOW()`,
		groupId, content)
	return err
}

// --- helpers ---

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
