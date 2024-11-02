package session

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/genvmoroz/lale/service/pkg/entity"
)

const defaultExpireIn = 24 * time.Hour

type (
	Repo struct {
		cache cache
		mux   *sync.Mutex
	}

	cache struct {
		entries  map[string]cacheEntry
		expireIn time.Duration
	}

	cacheEntry struct {
		session   entity.UserSession
		expiresAt time.Time
	}
)

var ErrOpenedSession = errors.New("session already opened") // todo: move to core layer

func NewRepo() (*Repo, error) {
	return &Repo{
		cache: cache{
			entries:  make(map[string]cacheEntry),
			expireIn: defaultExpireIn,
		},
		mux: &sync.Mutex{},
	}, nil
}

func (r *Repo) CreateSession(userID string) error {
	if !utf8.ValidString(userID) {
		return fmt.Errorf("invalid userID: %s", userID)
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	entry, exist := r.cache.get(userID)
	if exist && !entry.session.IsClosed() {
		return ErrOpenedSession
	}

	r.cache.set(entity.NewUserSession(userID))

	return nil
}

func (r *Repo) CloseSession(userID string) error {
	if !utf8.ValidString(userID) {
		return fmt.Errorf("invalid userID: %s", userID)
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	entry, exist := r.cache.get(userID)
	if !exist {
		return errors.New("session does not exist")
	}

	if entry.session.IsClosed() {
		return nil
	}

	entry.session.Close()

	r.cache.set(entry.session)

	return nil
}

func (c cache) get(userID string) (cacheEntry, bool) {
	if entry, ok := c.entries[userID]; ok && entry.expiresAt.After(time.Now()) {
		return entry, true
	}

	return cacheEntry{}, false
}

func (c cache) set(session entity.UserSession) {
	entry := cacheEntry{
		session:   session,
		expiresAt: time.Now().Add(c.expireIn),
	}

	c.entries[session.UserID] = entry
}
