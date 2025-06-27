package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestBruteForceLoginWithoutRateLimit(t *testing.T) {
	// Create test server with vulnerable login endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate login validation (vulnerable version without rate limiting)
		if r.Method == "POST" && r.URL.Path == "/login" {
			r.ParseForm()
			if r.Form.Get("username") == "admin" && r.Form.Get("password") == "correct_password" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Login successful"))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid credentials"))
			}
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	const (
		threadCount = 20        // Number of concurrent attackers
		attempts    = 10        // Attempts per attacker
		targetUser  = "admin"   // Target account
		password    = "pass123" // Common weak password
	)

	successCount := 0
	failureCount := 0
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go func(attackerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}

			for j := 0; j < attempts; j++ {
				// Craft brute force request
				body := bytes.NewBufferString(
					"username=" + targetUser + "&password=" + password,
				)

				resp, err := client.Post(
					server.URL+"/login",
					"application/x-www-form-urlencoded",
					body,
				)

				if err != nil {
					t.Logf("Attacker %d attempt %d failed: %v", attackerID, j, err)
					continue
				}

				if resp.StatusCode == http.StatusOK {
					successCount++
				} else {
					failureCount++
				}
				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("Brute force results:")
	t.Logf("Total attempts: %d", threadCount*attempts)
	t.Logf("Successful logins: %d", successCount)
	t.Logf("Failed logins: %d", failureCount)
	t.Logf("Attack duration: %v", duration)
	t.Logf("Requests/sec: %.1f", float64(threadCount*attempts)/duration.Seconds())

	// Vulnerability detection
	if successCount > 0 {
		t.Errorf("VULNERABILITY: System allowed %d successful brute force logins", successCount)
	} else if failureCount == 0 {
		t.Error("Test inconclusive - all requests failed")
	} else {
		t.Log("System appears to block brute force attacks")
	}
}
