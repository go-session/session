package session

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/tidwall/buntdb"
)

var (
	_ ManagerStore = &defaultManagerStore{}
	_ Store        = &defaultStore{}
)

// ManagerStore Management of session storage, including creation, update, and delete operations
type ManagerStore interface {
	// Create a session store and specify the expiration time (in seconds)
	Create(ctx context.Context, sid string, expired int64) (Store, error)
	// Update a session store and specify the expiration time (in seconds)
	Update(ctx context.Context, sid string, expired int64) (Store, error)
	// Delete a session store
	Delete(ctx context.Context, sid string) error
	// Check the session store exists
	Check(ctx context.Context, sid string) (bool, error)
	// Close storage, release resources
	Close() error
}

// Store A session id storage operation
type Store interface {
	// Get a session storage context
	Context() context.Context
	// Get the current session id
	SessionID() string
	// Set session value, call save function to take effect
	Set(key, value string)
	// Get session value
	Get(key string) (string, bool)
	// Delete session value, call save function to take effect
	Delete(key string) string
	// Clear all session data, call save function to take effect
	Flush()
	// Save session data
	Save() error
}

// NewMemoryStore Create an instance of a memory store
func NewMemoryStore() ManagerStore {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		panic(err)
	}

	return &defaultManagerStore{db}
}

// NewFileStore Create an instance of a file store
func NewFileStore(path string) ManagerStore {
	db, err := buntdb.Open(path)
	if err != nil {
		panic(err)
	}

	return &defaultManagerStore{db}
}

type defaultManagerStore struct {
	db *buntdb.DB
}

func (s *defaultManagerStore) getValue(sid string) (string, error) {
	var value string

	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(sid)
		if err != nil {
			if err == buntdb.ErrNotFound {
				return nil
			}
			return err
		}
		value = val
		return nil
	})

	return value, err
}

func (s *defaultManagerStore) parseValue(value string) (map[string]string, error) {
	var values map[string]string

	if len(value) > 0 {
		err := json.Unmarshal([]byte(value), &values)
		if err != nil {
			return nil, err
		}
	}

	if values == nil {
		values = make(map[string]string)
	}
	return values, nil
}

func (s *defaultManagerStore) Create(ctx context.Context, sid string, expired int64) (Store, error) {
	values := make(map[string]string)
	return &defaultStore{ctx: ctx, sid: sid, db: s.db, expired: expired, values: values}, nil
}

func (s *defaultManagerStore) Update(ctx context.Context, sid string, expired int64) (Store, error) {
	value, err := s.getValue(sid)
	if err != nil {
		return nil, err
	} else if value == "" {
		return s.Create(ctx, sid, expired)
	}

	err = s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err = tx.Set(sid, value, &buntdb.SetOptions{Expires: true, TTL: time.Duration(expired) * time.Second})
		return err
	})
	if err != nil {
		return nil, err
	}

	values, err := s.parseValue(value)
	if err != nil {
		return nil, err
	}

	return &defaultStore{ctx: ctx, sid: sid, db: s.db, expired: expired, values: values}, nil
}

func (s *defaultManagerStore) Delete(_ context.Context, sid string) error {
	return s.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(sid)
		if err == buntdb.ErrNotFound {
			return nil
		}
		return err
	})
}

func (s *defaultManagerStore) Check(_ context.Context, sid string) (bool, error) {
	var exists bool
	err := s.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(sid)
		if err != nil {
			if err == buntdb.ErrNotFound {
				return nil
			}
			return err
		}
		exists = true
		return nil
	})
	return exists, err
}

func (s *defaultManagerStore) Close() error {
	return s.db.Close()
}

type defaultStore struct {
	sid     string
	db      *buntdb.DB
	expired int64
	values  map[string]string
	sync.RWMutex
	ctx context.Context
}

func (s *defaultStore) Context() context.Context {
	return s.ctx
}

func (s *defaultStore) SessionID() string {
	return s.sid
}

func (s *defaultStore) Set(key, value string) {
	s.Lock()
	s.values[key] = value
	s.Unlock()
}

func (s *defaultStore) Get(key string) (string, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.values[key]
	return val, ok
}

func (s *defaultStore) Delete(key string) string {
	s.RLock()
	v, ok := s.values[key]
	s.RUnlock()
	if ok {
		s.Lock()
		delete(s.values, key)
		s.Unlock()
	}
	return v
}

func (s *defaultStore) Flush() {
	s.Lock()
	s.values = make(map[string]string)
	s.Unlock()
}

func (s *defaultStore) Save() error {
	var value string

	s.RLock()
	if len(s.values) > 0 {
		buf, _ := json.Marshal(s.values)
		value = string(buf)
	}
	s.RUnlock()

	return s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(s.sid, value, &buntdb.SetOptions{Expires: true, TTL: time.Duration(s.expired) * time.Second})
		return err
	})
}
