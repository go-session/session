package session

import (
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
func Start(w http.ResponseWriter, r *http.Request) (Store, error) {
	return defaultManager.Start(w, r)
}

// Destroy Destroy a session
func Destroy(w http.ResponseWriter, r *http.Request) error {
	return defaultManager.Destroy(w, r)
}
