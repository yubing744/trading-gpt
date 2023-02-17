package chat

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("chat", "ChatSessions")

type ChatSessions struct {
	sessions []*ChatSession
	lock     sync.RWMutex
}

func NewChatSessions() *ChatSessions {
	return &ChatSessions{
		sessions: make([]*ChatSession, 0),
	}
}

func (mgr *ChatSessions) AddChatSession(session *ChatSession) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	mgr.sessions = append(mgr.sessions, session)
}

func (mgr *ChatSessions) Notify(ctx context.Context, msg *types.Message) error {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()

	errs := make([]error, 0)

	for _, session := range mgr.sessions {
		err := session.Reply(ctx, msg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("Notify with many error, errors: %v", errs)
	}

	return nil
}
