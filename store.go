package session

import (
	"context"
	"sync"
	"time"

	"github.com/json-iterator/go"
	"github.com/tidwall/buntdb"
)

var (
	_             ManagerStore = &defaultManagerStore{}
	_             Store        = &defaultStore{}
	jsonMarshal                = jsoniter.Marshal
	jsonUnmarshal              = jsoniter.Unmarshal
)

// ManagerStore Management of session storage, including creation, update, and delete operations
type ManagerStore interface {
	// Check the session store exists
	Check(ctx context.Context, sid string) (bool, error)
	// Create a session store and specify the expiration time (in seconds)
	Create(ctx context.Context, sid string, expired int64) (Store, error)
	// Update a session store and specify the expiration time (in seconds)
	Update(ctx context.Context, sid string, expired int64) (Store, error)
	// Delete a session store
	Delete(ctx context.Context, sid string) error
	// Use sid to replace old sid and return session store
	Refresh(ctx context.Context, oldsid, sid string, expired int64) (Store, error)
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
	Set(key string, value interface{})
	// Get session value
	Get(key string) (interface{}, bool)
	// Delete session value, call save function to take effect
	Delete(key string) interface{}
	// Save session data
	Save() error
	// Clear all session data
	Flush() error
}

// NewMemoryStore Create an instance of a memory store
func NewMemoryStore() ManagerStore {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		panic(err)
	}
	return newDefaultManagerStore(db)
}

// NewFileStore Create an instance of a file store
func NewFileStore(path string) ManagerStore {
	db, err := buntdb.Open(path)
	if err != nil {
		panic(err)
	}
	return newDefaultManagerStore(db)
}

func newDefaultManagerStore(db *buntdb.DB) *defaultManagerStore {
	return &defaultManagerStore{
		db: db,
		pool: sync.Pool{
			New: func() interface{} {
				return newDefaultStore(db)
			},
		},
	}
}

type defaultManagerStore struct {
	db   *buntdb.DB
	pool sync.Pool
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

func (s *defaultManagerStore) parseValue(value string) (map[string]interface{}, error) {
	var values map[string]interface{}

	if len(value) > 0 {
		err := jsonUnmarshal([]byte(value), &values)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (s *defaultManagerStore) Check(_ context.Context, sid string) (bool, error) {
	val, err := s.getValue(sid)
	if err != nil {
		return false, err
	}
	return val != "", nil
}

func (s *defaultManagerStore) Create(ctx context.Context, sid string, expired int64) (Store, error) {
	store := s.pool.Get().(*defaultStore)
	store.reset(ctx, sid, expired, nil)
	return store, nil
}

func (s *defaultManagerStore) Update(ctx context.Context, sid string, expired int64) (Store, error) {
	store := s.pool.Get().(*defaultStore)

	value, err := s.getValue(sid)
	if err != nil {
		return nil, err
	} else if value == "" {
		store.reset(ctx, sid, expired, nil)
		return store, nil
	}

	err = s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(sid, value,
			&buntdb.SetOptions{Expires: true, TTL: time.Duration(expired) * time.Second})
		return err
	})
	if err != nil {
		return nil, err
	}

	values, err := s.parseValue(value)
	if err != nil {
		return nil, err
	}

	store.reset(ctx, sid, expired, values)
	return store, nil
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

func (s *defaultManagerStore) Refresh(ctx context.Context, oldsid, sid string, expired int64) (Store, error) {
	store := s.pool.Get().(*defaultStore)

	value, err := s.getValue(oldsid)
	if err != nil {
		return nil, err
	} else if value == "" {
		store.reset(ctx, sid, expired, nil)
		return store, nil
	}

	err = s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(sid, value,
			&buntdb.SetOptions{Expires: true, TTL: time.Duration(expired) * time.Second})
		if err != nil {
			return err
		}
		_, err = tx.Delete(oldsid)
		return err
	})
	if err != nil {
		return nil, err
	}

	values, err := s.parseValue(value)
	if err != nil {
		return nil, err
	}

	store.reset(ctx, sid, expired, values)
	return store, nil
}

func (s *defaultManagerStore) Close() error {
	return s.db.Close()
}

func newDefaultStore(db *buntdb.DB) *defaultStore {
	return &defaultStore{
		db: db,
	}
}

type defaultStore struct {
	sync.RWMutex
	ctx     context.Context
	sid     string
	expired int64
	db      *buntdb.DB
	values  map[string]interface{}
}

func (s *defaultStore) reset(ctx context.Context, sid string, expired int64, values map[string]interface{}) {
	if values == nil {
		values = make(map[string]interface{})
	}
	s.ctx = ctx
	s.sid = sid
	s.expired = expired
	s.values = values
}

func (s *defaultStore) Context() context.Context {
	return s.ctx
}

func (s *defaultStore) SessionID() string {
	return s.sid
}

func (s *defaultStore) Set(key string, value interface{}) {
	s.Lock()
	s.values[key] = value
	s.Unlock()
}

func (s *defaultStore) Get(key string) (interface{}, bool) {
	s.RLock()
	val, ok := s.values[key]
	s.RUnlock()
	return val, ok
}

func (s *defaultStore) Delete(key string) interface{} {
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

func (s *defaultStore) Flush() error {
	s.Lock()
	s.values = make(map[string]interface{})
	s.Unlock()
	return s.Save()
}

func (s *defaultStore) Save() error {
	var value string

	s.RLock()
	if len(s.values) > 0 {
		buf, err := jsonMarshal(s.values)
		if err != nil {
			s.RUnlock()
			return err
		}
		value = string(buf)
	}
	s.RUnlock()

	return s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(s.sid, value,
			&buntdb.SetOptions{Expires: true, TTL: time.Duration(s.expired) * time.Second})
		return err
	})
}
