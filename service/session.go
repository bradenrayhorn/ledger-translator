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

func (s Session) GetSession(sessionID string) (string, error) {
	r, err := s.client.Authenticate(context.Background(), &session.SessionAuthenticateRequest{
		SessionID: sessionID,
		UserAgent: "test",
		IP:        "127.0.0.1",
	})
	if err != nil {
		return "", err
	}

	return r.Session.GetUserID(), nil
}
