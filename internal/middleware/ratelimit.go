package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// loginAttempt tracks failed login attempts per IP for rate limiting.
type ipEntry struct {
	count    int
	windowStart time.Time
}

var (
	loginMu     sync.Mutex
	loginTracker = make(map[string]*ipEntry)
)

// LoginRateLimit allows at most maxAttempts failed login attempts per IP
// within the window duration. After that, the IP is blocked for the window.
func LoginRateLimit() gin.HandlerFunc {
	const maxAttempts = 10
	const window = 5 * time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()

		loginMu.Lock()
		entry, exists := loginTracker[ip]
		now := time.Now()

		if !exists || now.Sub(entry.windowStart) > window {
			loginTracker[ip] = &ipEntry{count: 1, windowStart: now}
			loginMu.Unlock()
			c.Next()
			return
		}

		if entry.count >= maxAttempts {
			loginMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "登录尝试过于频繁，请5分钟后再试"})
			c.Abort()
			return
		}

		entry.count++
		loginMu.Unlock()

		c.Next()

		// If login succeeded (status 200), reset the counter
		if c.Writer.Status() == http.StatusOK {
			loginMu.Lock()
			delete(loginTracker, ip)
			loginMu.Unlock()
		}
	}
}
