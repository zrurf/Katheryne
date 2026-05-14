package dao

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserDBDao struct {
	db *pgxpool.Pool
}

func NewUserDBDao(pool *pgxpool.Pool) *UserDBDao {
	return &UserDBDao{db: pool}
}

func (d *UserDBDao) GetUserById(ctx context.Context, uid int64) (*User, error) {
	row := d.db.QueryRow(ctx,
		`SELECT uid, name, phone, avatar, status, created_at, updated_at, last_login FROM users WHERE uid = $1`, uid)
	u := &User{}
	err := row.Scan(&u.Uid, &u.Name, &u.Phone, &u.Avatar, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.LastLogin)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (d *UserDBDao) SearchUser(ctx context.Context, keyword string, page, size int32) ([]*User, int64, error) {
	if size <= 0 || size > 100 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * size

	like := "%" + keyword + "%"

	var total int64
	err := d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE name ILIKE $1 OR phone ILIKE $1`, like).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := d.db.Query(ctx,
		`SELECT uid, name, phone, avatar, status, created_at, updated_at, last_login FROM users
		 WHERE name ILIKE $1 OR phone ILIKE $1
		 ORDER BY uid DESC LIMIT $2 OFFSET $3`,
		like, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*User
	for rows.Next() {
		u := &User{}
		err := rows.Scan(&u.Uid, &u.Name, &u.Phone, &u.Avatar, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.LastLogin)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, u)
	}
	return list, total, rows.Err()
}

func (d *UserDBDao) UpdateUserInfo(ctx context.Context, uid int64, nickname, avatar, gender, bio string) error {
	var setParts []string
	var args []interface{}
	argIdx := 1

	if nickname != "" {
		setParts = append(setParts, "name = $"+strconv.Itoa(argIdx))
		args = append(args, nickname)
		argIdx++
	}
	if avatar != "" {
		setParts = append(setParts, "avatar = $"+strconv.Itoa(argIdx))
		args = append(args, avatar)
		argIdx++
	}
	if len(setParts) == 0 {
		return nil
	}

	args = append(args, uid)
	query := "UPDATE users SET " + strings.Join(setParts, ", ") + ", updated_at = NOW() WHERE uid = $" + strconv.Itoa(argIdx)
	_, err := d.db.Exec(ctx, query, args...)
	return err
}
