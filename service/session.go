package service

import (
	"context"

	"github.com/bradenrayhorn/ledger-protos/session"
)

type Session struct {
	client session.SessionAuthenticatorClient
}

func NewSessionService(client session.SessionAuthenticatorClient) Session {
	return Session{
		client: client,
	}
}

func (s Session) GetSession(sessionID string, ip string, userAgent string) (string, error) {
	r, err := s.client.Authenticate(context.Background(), &session.SessionAuthenticateRequest{
		SessionID: sessionID,
		UserAgent: userAgent,
		IP:        ip,
	})
	if err != nil {
		return "", err
	}

	return r.Session.GetUserID(), nil
}
