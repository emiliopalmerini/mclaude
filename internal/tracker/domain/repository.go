package domain

// SessionRepository defines the interface for persisting sessions
type SessionRepository interface {
	Save(session Session) error
}
