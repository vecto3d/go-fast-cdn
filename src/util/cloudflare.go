package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Cloudflare accepts at most 30 URLs per purge call on the plans this runs on.
const purgeBatchSize = 30

var purgeClient = &http.Client{Timeout: 10 * time.Second}

// PurgeConfigured reports whether cache purging can run. Without both values
// the CDN still works; replaced files just serve from cache until their TTL
// expires, which is how it behaved before.
func PurgeConfigured() bool {
	return os.Getenv("CLOUDFLARE_API_TOKEN") != "" && os.Getenv("CLOUDFLARE_ZONE_ID") != ""
}

// PublicBaseURL is the origin the CDN is served from, used to build the
// absolute URLs Cloudflare purges by. Falls back to the request host when unset.
func PublicBaseURL(requestHost string) string {
	if base := os.Getenv("PUBLIC_BASE_URL"); base != "" {
		return strings.TrimSuffix(base, "/")
	}
	if requestHost == "" {
		return ""
	}

	return "https://" + requestHost
}

// PurgeURLs drops the given URLs from Cloudflare's edge cache.
//
// Files are served with a day-long max-age, so replacing or deleting one leaves
// the old bytes being served from the edge until that expires. Purging on
// change is what makes an updated file actually update for everyone.
//
// It runs in the background: a failed purge should never fail the upload or
// delete that triggered it, and the caller has already done the real work.
func PurgeURLs(urls []string) {
	if len(urls) == 0 || !PurgeConfigured() {
		return
	}

	go func() {
		for start := 0; start < len(urls); start += purgeBatchSize {
			end := start + purgeBatchSize
			if end > len(urls) {
				end = len(urls)
			}

			if err := purgeBatch(urls[start:end]); err != nil {
				log.Printf("Cloudflare purge failed for %d urls: %s", end-start, err)
			}
		}
	}()
}

func purgeBatch(urls []string) error {
	payload, err := json.Marshal(map[string]any{"files": urls})
	if err != nil {
		return err
	}

	return postPurge(payload)
}

// apiBase is the Cloudflare API root, overridable so the call can be pointed at
// a stand-in during tests.
func apiBase() string {
	if base := os.Getenv("CLOUDFLARE_API_BASE"); base != "" {
		return strings.TrimSuffix(base, "/")
	}

	return "https://api.cloudflare.com/client/v4"
}

func postPurge(payload []byte) error {
	endpoint := apiBase() + "/zones/" + os.Getenv("CLOUDFLARE_ZONE_ID") + "/purge_cache"

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+os.Getenv("CLOUDFLARE_API_TOKEN"))
	request.Header.Set("Content-Type", "application/json")

	response, err := purgeClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body := make([]byte, 512)
		read, _ := response.Body.Read(body)
		return &purgeError{status: response.StatusCode, body: string(body[:read])}
	}

	return nil
}

type purgeError struct {
	status int
	body   string
}

func (e *purgeError) Error() string {
	return "cloudflare responded " + http.StatusText(e.status) + ": " + e.body
}

// PurgeEverything drops the whole zone from Cloudflare's cache. Blunt, but it
// is the only way to clear entries written before purging was configured, and
// the only way to catch URLs the app does not know it served.
func PurgeEverything() error {
	if !PurgeConfigured() {
		return &purgeError{status: http.StatusPreconditionFailed, body: "purging is not configured"}
	}

	payload, err := json.Marshal(map[string]any{"purge_everything": true})
	if err != nil {
		return err
	}

	return postPurge(payload)
}
