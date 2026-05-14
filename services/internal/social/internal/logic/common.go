package logic

import "database/sql"

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
