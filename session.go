package session

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

var (
	// ErrInvalidSessionID invalid session id
	ErrInvalidSessionID = errors.New("invalid session id")
)

// Define default options
var defaultOptions = options{
	cookieName:     "session_id",
	cookieLifeTime: 3600 * 24,
	expired:        7200,
	store:          NewMemoryStore(),
	sessionID: func() string {
		return uuid.Must(uuid.NewV4()).String()
	},
}

type options struct {
	sign           []byte
	cookieName     string
	cookieLifeTime int
	secure         bool
	domain         string
	expired        int64
	sessionID      func() string
	store          ManagerStore
}

// Option A session parameter options
type Option func(*options)

// SetSign Set the session id signature value
func SetSign(sign []byte) Option {
	return func(o *options) {
		o.sign = sign
	}
}

// SetCookieName Set the cookie name
func SetCookieName(cookieName string) Option {
	return func(o *options) {
		o.cookieName = cookieName
	}
}

// SetCookieLifeTime Set the cookie expiration time (in seconds)
func SetCookieLifeTime(cookieLifeTime int) Option {
	return func(o *options) {
		o.cookieLifeTime = cookieLifeTime
	}
}

// SetDomain Set the domain name of the cookie
func SetDomain(domain string) Option {
	return func(o *options) {
		o.domain = domain
	}
}

// SetSecure Set cookie security
func SetSecure(secure bool) Option {
	return func(o *options) {
		o.secure = secure
	}
}

// SetExpired Set session expiration time (in seconds)
func SetExpired(expired int64) Option {
	return func(o *options) {
		o.expired = expired
	}
}

// SetSessionID Set callback function to generate session id
func SetSessionID(sessionID func() string) Option {
	return func(o *options) {
		o.sessionID = sessionID
	}
}

// SetStore Set session management storage
func SetStore(store ManagerStore) Option {
	return func(o *options) {
		o.store = store
	}
}

// NewManager Create a session management instance
func NewManager(opt ...Option) *Manager {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	if opts.store == nil {
		panic("unknown store")
	}
	return &Manager{opts: opts}
}

// Manager A session management instance, including start and destroy operations
type Manager struct {
	opts options
}

// Start Start a session and return to session storage
func (m *Manager) Start(ctx context.Context, w http.ResponseWriter, r *http.Request) (Store, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = newReqContext(ctx, r)
	ctx = newResContext(ctx, w)

	sid, err := m.sessionID(r)
	if err != nil {
		return nil, err
	}

	if sid != "" {
		if exists, verr := m.opts.store.Check(ctx, sid); verr != nil {
			return nil, verr
		} else if exists {
			return m.opts.store.Update(ctx, sid, m.opts.expired)
		}
	}

	store, err := m.opts.store.Create(ctx, m.opts.sessionID(), m.opts.expired)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     m.opts.cookieName,
		Value:    m.encodeSessionID(store.SessionID()),
		Path:     "/",
		HttpOnly: true,
		Secure:   m.isSecure(r),
		Domain:   m.opts.domain,
	}

	if v := m.opts.cookieLifeTime; v > 0 {
		cookie.MaxAge = v
		cookie.Expires = time.Now().Add(time.Duration(v) * time.Second)
	}

	http.SetCookie(w, cookie)
	r.AddCookie(cookie)

	return store, nil
}

// Destroy Destroy a session
func (m *Manager) Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = newReqContext(ctx, r)
	ctx = newResContext(ctx, w)

	sid, err := m.sessionID(r)
	if err != nil {
		return err
	} else if sid == "" {
		return nil
	}

	err = m.opts.store.Delete(ctx, sid)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     m.opts.cookieName,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
		MaxAge:   -1,
	}

	http.SetCookie(w, cookie)
	return nil
}

func (m *Manager) signature(sid string) string {
	h := sha1.New()
	h.Write(m.opts.sign)
	h.Write([]byte(sid))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func (m *Manager) encodeSessionID(sid string) string {
	b := base64.StdEncoding.EncodeToString([]byte(sid))
	s := fmt.Sprintf("%s.%s", b, m.signature(sid))
	return url.QueryEscape(s)
}

func (m *Manager) decodeSessionID(value string) (string, error) {
	value, err := url.QueryUnescape(value)
	if err != nil {
		return "", err
	}

	vals := strings.Split(value, ".")
	if len(vals) != 2 {
		return "", ErrInvalidSessionID
	}

	bsid, err := base64.StdEncoding.DecodeString(vals[0])
	if err != nil {
		return "", err
	}
	sid := string(bsid)

	sign := m.signature(sid)
	if sign != vals[1] {
		return "", ErrInvalidSessionID
	}
	return sid, nil
}

func (m *Manager) sessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(m.opts.cookieName)
	if err == nil && cookie.Value != "" {
		return m.decodeSessionID(cookie.Value)
	}
	return "", nil
}

func (m *Manager) isSecure(r *http.Request) bool {
	if !m.opts.secure {
		return false
	}
	if r.URL.Scheme != "" {
		return r.URL.Scheme == "https"
	}
	if r.TLS == nil {
		return false
	}
	return true
}
