package util

import (
	"bytes"
	"net/http"
)

// flacMagic is the marker every FLAC stream starts with.
var flacMagic = []byte("fLaC")

// DetectContentType wraps http.DetectContentType and fills in the formats its
// sniff table does not cover. FLAC is the only common upload here that Go
// reports as application/octet-stream, which would make it indistinguishable
// from an arbitrary binary and get it rejected.
func DetectContentType(buffer []byte) string {
	if bytes.HasPrefix(buffer, flacMagic) {
		return "audio/flac"
	}

	return http.DetectContentType(buffer)
}
