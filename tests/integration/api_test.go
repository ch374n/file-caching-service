package integration_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestGetFileAPI tests the GET /files/{name} endpoint against a running service
// This test requires:
// - The service to be running and accessible
// - R2 storage to have test files
// - SERVICE_URL environment variable set (e.g., http://localhost:8080)
func TestGetFileAPI(t *testing.T) {
	serviceURL := os.Getenv("SERVICE_URL")
	if serviceURL == "" {
		t.Skip("SERVICE_URL not set, skipping integration test")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.Get(serviceURL + "/health")
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Health check returned %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		if result["success"] != true {
			t.Errorf("Health check not successful: %v", result)
		}
	})

	t.Run("Get Existing File", func(t *testing.T) {
		testFile := os.Getenv("TEST_FILE_NAME")
		if testFile == "" {
			testFile = "test.txt" // Default test file
		}

		resp, err := client.Get(serviceURL + "/files/" + testFile)
		if err != nil {
			t.Fatalf("GET file failed: %v", err)
		}
		defer resp.Body.Close()

		// File might exist or not - we're testing the endpoint works
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Unexpected status %d: %s", resp.StatusCode, string(body))
		}

		if resp.StatusCode == http.StatusOK {
			// Verify we got content
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			if len(body) == 0 {
				t.Error("Expected non-empty file content")
			}
			t.Logf("Successfully retrieved file: %s (%d bytes)", testFile, len(body))
		} else {
			t.Logf("File not found (expected if test file doesn't exist): %s", testFile)
		}
	})

	t.Run("Get Non-Existent File", func(t *testing.T) {
		resp, err := client.Get(serviceURL + "/files/non-existent-file-12345.txt")
		if err != nil {
			t.Fatalf("GET file failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for non-existent file, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["success"] != false {
			t.Errorf("Expected success=false for non-existent file")
		}
	})

	t.Run("Cache Behavior", func(t *testing.T) {
		testFile := os.Getenv("TEST_FILE_NAME")
		if testFile == "" {
			t.Skip("TEST_FILE_NAME not set, skipping cache test")
		}

		// First request - should be cache miss
		start1 := time.Now()
		resp1, err := client.Get(serviceURL + "/files/" + testFile)
		if err != nil {
			t.Fatalf("First request failed: %v", err)
		}
		duration1 := time.Since(start1)
		resp1.Body.Close()

		if resp1.StatusCode != http.StatusOK {
			t.Skip("Test file not found, skipping cache test")
		}

		// Wait a bit for async caching
		time.Sleep(500 * time.Millisecond)

		// Second request - should be cache hit (faster)
		start2 := time.Now()
		resp2, err := client.Get(serviceURL + "/files/" + testFile)
		if err != nil {
			t.Fatalf("Second request failed: %v", err)
		}
		duration2 := time.Since(start2)
		resp2.Body.Close()

		t.Logf("First request: %v, Second request: %v", duration1, duration2)

		// Cache hit should generally be faster, but we can't guarantee it
		// Just log the times for manual verification
	})

	t.Run("Content-Type Detection", func(t *testing.T) {
		testCases := []struct {
			filename    string
			contentType string
		}{
			{"test.pdf", "application/pdf"},
			{"test.html", "text/html"},
			{"test.json", "application/json"},
			{"test.txt", "text/plain"},
		}

		for _, tc := range testCases {
			t.Run(tc.filename, func(t *testing.T) {
				resp, err := client.Get(serviceURL + "/files/" + tc.filename)
				if err != nil {
					t.Fatalf("GET failed: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					contentType := resp.Header.Get("Content-Type")
					// Just log - file might not exist
					t.Logf("File %s: Content-Type=%s", tc.filename, contentType)
				}
			})
		}
	})

	t.Run("Root Endpoint", func(t *testing.T) {
		resp, err := client.Get(serviceURL + "/")
		if err != nil {
			t.Fatalf("GET / failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result["message"] != "File Caching Service" {
			t.Errorf("Unexpected message: %v", result["message"])
		}
	})

	t.Run("Metrics Endpoint", func(t *testing.T) {
		resp, err := client.Get(serviceURL + "/metrics")
		if err != nil {
			t.Fatalf("GET /metrics failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Check for expected metrics
		expectedMetrics := []string{
			"http_requests_total",
			"cache_hits_total",
			"cache_misses_total",
		}

		for _, metric := range expectedMetrics {
			if !strings.Contains(bodyStr, metric) {
				t.Errorf("Expected metric %s not found", metric)
			}
		}
	})
}

// TestServiceAvailability is a simple smoke test
func TestServiceAvailability(t *testing.T) {
	serviceURL := os.Getenv("SERVICE_URL")
	if serviceURL == "" {
		t.Skip("SERVICE_URL not set")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Retry a few times in case service is still starting
	var lastErr error
	for i := 0; i < 30; i++ {
		resp, err := client.Get(serviceURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Log("Service is available")
			return
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("status code: %d", resp.StatusCode)
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Service not available after 60 seconds: %v", lastErr)
}
