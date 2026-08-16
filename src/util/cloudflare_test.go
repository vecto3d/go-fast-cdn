package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPurgeConfiguredNeedsBothValues(t *testing.T) {
	tests := []struct {
		name  string
		token string
		zone  string
		want  bool
	}{
		{"neither", "", "", false},
		{"token only", "secret", "", false},
		{"zone only", "", "zone123", false},
		{"both", "secret", "zone123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLOUDFLARE_API_TOKEN", tt.token)
			t.Setenv("CLOUDFLARE_ZONE_ID", tt.zone)

			require.Equal(t, tt.want, PurgeConfigured())
		})
	}
}

// Purging is opt-in: an install without credentials must behave exactly as it
// did before, never reaching out to anything.
func TestPurgeIsANoOpWithoutCredentials(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "")
	t.Setenv("CLOUDFLARE_ZONE_ID", "")

	require.NotPanics(t, func() {
		PurgeURLs([]string{"https://cdn.example.com/api/cdn/download/images/a.png"})
	})
	require.Error(t, PurgeEverything())
}

func TestPublicBaseURL(t *testing.T) {
	t.Run("configured value wins and loses its trailing slash", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "https://cdn.example.com/")

		require.Equal(t, "https://cdn.example.com", PublicBaseURL("ignored-host"))
	})

	t.Run("falls back to the request host", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "")

		require.Equal(t, "https://cdn.example.com", PublicBaseURL("cdn.example.com"))
	})

	t.Run("no host and no config yields nothing to purge", func(t *testing.T) {
		t.Setenv("PUBLIC_BASE_URL", "")

		require.Equal(t, "", PublicBaseURL(""))
	})
}

// The request Cloudflare receives has to carry the token and the exact URLs, so
// this points the client at a stand-in and inspects what arrives.
func TestPurgeSendsTheURLsWithTheToken(t *testing.T) {
	type received struct {
		path  string
		auth  string
		files []string
	}

	got := make(chan received, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Files []string `json:"files"`
		}{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		got <- received{path: r.URL.Path, auth: r.Header.Get("Authorization"), files: body.Files}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
	t.Setenv("CLOUDFLARE_ZONE_ID", "zone123")
	t.Setenv("CLOUDFLARE_API_BASE", server.URL)

	PurgeURLs([]string{
		"https://cdn.example.com/api/cdn/download/images/logos/a.png",
		"https://cdn.example.com/api/cdn/download/images/logos/b.png",
	})

	select {
	case call := <-got:
		require.Equal(t, "/zones/zone123/purge_cache", call.path)
		require.Equal(t, "Bearer test-token", call.auth)
		require.Equal(t, []string{
			"https://cdn.example.com/api/cdn/download/images/logos/a.png",
			"https://cdn.example.com/api/cdn/download/images/logos/b.png",
		}, call.files)
	case <-time.After(3 * time.Second):
		t.Fatal("no purge request arrived")
	}
}

// Cloudflare caps a purge at 30 urls, so a larger set has to be split.
func TestPurgeSplitsLargeSetsIntoBatches(t *testing.T) {
	batches := make(chan int, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Files []string `json:"files"`
		}{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		batches <- len(body.Files)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
	t.Setenv("CLOUDFLARE_ZONE_ID", "zone123")
	t.Setenv("CLOUDFLARE_API_BASE", server.URL)

	urls := make([]string, 70)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://cdn.example.com/api/cdn/download/images/%d.png", i)
	}
	PurgeURLs(urls)

	sizes := []int{}
	for len(sizes) < 3 {
		select {
		case size := <-batches:
			sizes = append(sizes, size)
		case <-time.After(3 * time.Second):
			t.Fatalf("only %d batches arrived: %v", len(sizes), sizes)
		}
	}

	require.Equal(t, []int{30, 30, 10}, sizes)
}
