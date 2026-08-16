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
// The cache window is a day rather than immutable, so replacing a file under
// the same name corrects itself without a purge.
func ServedFileHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "public, max-age=86400")
		c.Next()
	}
}
