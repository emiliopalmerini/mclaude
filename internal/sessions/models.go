package sessions

import "claude-watcher/internal/database/sqlc"

type SessionsData struct {
	Sessions   []sqlc.ListSessionsRow
	Page       int
	TotalPages int
}
