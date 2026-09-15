package middleware

import "github.com/gin-gonic/gin"

// ServedFileHeaders guards and caches the files the CDN serves.
//
// Uploaded content is served from the same origin as the dashboard, so an SVG —
// which is markup and can carry script — would otherwise run in that origin.
// The sandbox policy stops any script, plugin or form in a served file from
// executing, while leaving images, audio and video embedding untouched, and
// nosniff keeps a browser from re-interpreting a file as something it is not.
//
// Cloudflare keeps a file for a day (s-maxage) and is purged whenever one
// changes, but a purge can't reach browser caches, so browsers only keep it
// for five minutes. Otherwise a replaced file stays stale for players for a day.
func ServedFileHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "public, max-age=300, s-maxage=86400")
		c.Next()
	}
}
