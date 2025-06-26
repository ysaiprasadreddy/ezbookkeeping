package api

import (
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// resetLoginAttempts clears all IP tracking for a clean slate.
func resetLoginAttempts() {
	loginMutex.Lock()
	defer loginMutex.Unlock()
	loginAttempts = make(map[string][]time.Time)
}

func TestRateLimit_AllCases(t *testing.T) {
	ip1 := "192.168.0.1"
	ip2 := "192.168.0.2"

	// Common setup
	settings.LoginRateWindowSeconds = 2 // 2 seconds window for quick tests

	t.Run("Case 1: First request not rate limited", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 3
		if isRateLimited(ip1) {
			t.Errorf("Expected first request not to be rate limited")
		}
	})

	t.Run("Case 2: Multiple requests within limit", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 3
		for i := 0; i < 3; i++ {
			if isRateLimited(ip1) {
				t.Errorf("Attempt %d: unexpected rate limit", i+1)
			}
		}
	})

	t.Run("Case 3: Exactly at limit should still be allowed", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 3
		for i := 0; i < 2; i++ {
			isRateLimited(ip1) // warmup
		}
		if isRateLimited(ip1) != false {
			t.Errorf("3rd attempt should still be allowed")
		}
	})

	t.Run("Case 4: Over the limit should be blocked", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 2
		isRateLimited(ip1)
		isRateLimited(ip1)
		if !isRateLimited(ip1) {
			t.Errorf("Expected to be rate limited on 3rd attempt")
		}
	})

	t.Run("Case 5: Allow after window expiry", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 2
		isRateLimited(ip1)
		isRateLimited(ip1)
		time.Sleep(time.Duration(settings.LoginRateWindowSeconds+1) * time.Second)
		if isRateLimited(ip1) {
			t.Errorf("Expected request after window to be allowed")
		}
	})

	t.Run("Case 6: Different IP is tracked separately", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 1
		isRateLimited(ip1)
		if isRateLimited(ip2) {
			t.Errorf("Different IP should not be affected by other IP's limit")
		}
	})

	t.Run("Case 7: Zero limit should always block", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 0
		if !isRateLimited(ip1) {
			t.Errorf("With 0 limit, request should be denied")
		}
	})

	t.Run("Case 8: Large limit should not block", func(t *testing.T) {
		resetLoginAttempts()
		settings.LoginRateLimit = 1000
		blocked := false
		for i := 0; i < 100; i++ {
			if isRateLimited(ip1) {
				blocked = true
				break
			}
		}
		if blocked {
			t.Errorf("Should not be blocked within 100 attempts for large limit")
		}
	})
}
