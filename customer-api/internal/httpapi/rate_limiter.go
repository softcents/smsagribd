package httpapi

import (
	"sync"
	"time"
)

type loginAttemptInfo struct {
	attempts    int
	windowStart time.Time
	lockUntil   time.Time
}

// LoginLimiter provides in-memory thread-safe rate limiting for logins,
// matching Laravel's RateLimiter behavior (5 attempts per minute).
type LoginLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttemptInfo
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{
		attempts: make(map[string]*loginAttemptInfo),
	}
}

// Check checks if the key is currently locked out. If locked, returns (true, retryAfterSeconds).
func (l *LoginLimiter) Check(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	info, exists := l.attempts[key]
	if !exists {
		return false, 0
	}

	if now.Before(info.lockUntil) {
		remaining := int(info.lockUntil.Sub(now).Seconds()) + 1
		return true, remaining
	}

	// If lockout expired or window expired (60s), reset
	if now.Sub(info.windowStart) > time.Minute && !now.Before(info.lockUntil) {
		delete(l.attempts, key)
		return false, 0
	}

	return false, 0
}

// RecordFailure records a failed login attempt. If attempts reach 5, locks for 60 seconds.
func (l *LoginLimiter) RecordFailure(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	info, exists := l.attempts[key]
	if !exists || now.Sub(info.windowStart) > time.Minute {
		info = &loginAttemptInfo{
			attempts:    1,
			windowStart: now,
		}
		l.attempts[key] = info
		return false, 0
	}

	info.attempts++
	if info.attempts >= 5 {
		info.lockUntil = now.Add(time.Minute)
		return true, 60
	}

	return false, 0
}

// Reset clears attempts on successful login.
func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
