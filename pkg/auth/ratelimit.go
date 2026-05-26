package auth

import (
	"net"
	"sync"
	"time"
)

type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]*limitEntry
	limit   int
	window  time.Duration
}

type limitEntry struct {
	count   int
	resetAt time.Time
}

func newLoginLimiter(limit int, window time.Duration) *loginLimiter {
	return &loginLimiter{
		entries: make(map[string]*limitEntry),
		limit:   limit,
		window:  window,
	}
}

func (l *loginLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[key]
	if !ok || now.After(entry.resetAt) {
		l.entries[key] = &limitEntry{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if entry.count >= l.limit {
		return false
	}
	entry.count++
	return true
}

func requestClientKey(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
