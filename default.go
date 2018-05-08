package session

import (
	"context"
	"net/http"
)

var defaultManager = NewManager()

// InitManager Initialize the global session management instance
func InitManager(opt ...Option) {
	for _, o := range opt {
		o(&defaultManager.opts)
	}
}

// Start Start a session and return to session storage
func Start(ctx context.Context, w http.ResponseWriter, r *http.Request) (Store, error) {
	return defaultManager.Start(ctx, w, r)
}

// Destroy Destroy a session
func Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return defaultManager.Destroy(ctx, w, r)
}

// Refresh a session and return to session storage
func Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (Store, error) {
	return defaultManager.Refresh(ctx, w, r)
}
