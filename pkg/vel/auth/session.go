package auth

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
)

type Session struct {
	s     *sessions.Session
	store sessions.Store

	r *http.Request
	w http.ResponseWriter

	l *slog.Logger
}

func NewSession(store sessions.Store, r *http.Request, w http.ResponseWriter, l *slog.Logger) *Session {
	return &Session{nil, store, r, w, l}
}

func (s *Session) Session() *sessions.Session {
	if s.s == nil {
		var err error
		s.s, err = s.store.Get(s.r, sessionName)
		if err != nil {
			s.l.ErrorContext(s.r.Context(), "failed to get session", "err", err)
		}
	}

	return s.s
}

func (s *Session) Save() {
	if err := s.store.Save(s.r, s.w, s.Session()); err != nil {
		s.l.ErrorContext(s.r.Context(), "failed to save session", "err", err)
	}
}

const sessionName = "auth-session"

type SessionStore struct {
	session *Session
}

func (s *SessionStore) GetItem(key string) string {
	v := s.session.Session().Values[key]
	if v == nil {
		return ""
	}
	return v.(string)
}

func (s *SessionStore) SetItem(key, value string) {
	s.session.Session().Values[key] = value
	s.session.Session().Save(s.session.r, s.session.w)
}
