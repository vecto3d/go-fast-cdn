package util

import (
	"bytes"
	"path/filepath"
	"strings"
)

// part is one byte marker a format's content must carry. A non-negative offset
// means the marker sits exactly there; anywhere means it only has to appear
// somewhere in the sniffed prefix, which is how text formats like SVG have to
// be matched since they can start with a BOM, whitespace or a comment.
type part struct {
	offset   int
	anywhere bool
	magic    []byte
}

// Every part of a candidate must match; any one candidate matching is enough.
// RIFF containers need this: "RIFF" alone does not separate a .webp from a
// .wav or an .avi, only the tag at offset 8 does.
type candidate []part

var signatures = map[string][]candidate{
	// Images
	".png":  {{{offset: 0, magic: []byte("\x89PNG\r\n\x1a\n")}}},
	".jpg":  {{{offset: 0, magic: []byte{0xFF, 0xD8, 0xFF}}}},
	".jpeg": {{{offset: 0, magic: []byte{0xFF, 0xD8, 0xFF}}}},
	".gif":  {{{offset: 0, magic: []byte("GIF87a")}}, {{offset: 0, magic: []byte("GIF89a")}}},
	".webp": {{{offset: 0, magic: []byte("RIFF")}, {offset: 8, magic: []byte("WEBP")}}},
	".bmp":  {{{offset: 0, magic: []byte("BM")}}},
	".tif":  {{{offset: 0, magic: []byte("II*\x00")}}, {{offset: 0, magic: []byte("MM\x00*")}}},
	".tiff": {{{offset: 0, magic: []byte("II*\x00")}}, {{offset: 0, magic: []byte("MM\x00*")}}},
	".ico":  {{{offset: 0, magic: []byte{0x00, 0x00, 0x01, 0x00}}}},
	".avif": {{{offset: 4, magic: []byte("ftyp")}, {offset: 8, magic: []byte("avif")}}, {{offset: 4, magic: []byte("ftyp")}, {offset: 8, magic: []byte("avis")}}},
	".heic": {{{offset: 4, magic: []byte("ftyp")}, {offset: 8, magic: []byte("heic")}}, {{offset: 4, magic: []byte("ftyp")}, {offset: 8, magic: []byte("heix")}}, {{offset: 4, magic: []byte("ftyp")}, {offset: 8, magic: []byte("mif1")}}},
	".heif": {{{offset: 4, magic: []byte("ftyp")}, {offset: 8, magic: []byte("mif1")}}, {{offset: 4, magic: []byte("ftyp")}, {offset: 8, magic: []byte("heic")}}},
	".svg":  {{{anywhere: true, magic: []byte("<svg")}}},

	// Documents
	".pdf":  {{{offset: 0, magic: []byte("%PDF-")}}},
	".zip":  {{{offset: 0, magic: []byte("PK\x03\x04")}}},
	".docx": {{{offset: 0, magic: []byte("PK\x03\x04")}}},
	".xlsx": {{{offset: 0, magic: []byte("PK\x03\x04")}}},
	".pptx": {{{offset: 0, magic: []byte("PK\x03\x04")}}},
	".rtf":  {{{offset: 0, magic: []byte("{\\rtf")}}},

	// Audio
	".mp3":  {{{offset: 0, magic: []byte("ID3")}}, {{offset: 0, magic: []byte{0xFF, 0xFB}}}, {{offset: 0, magic: []byte{0xFF, 0xF3}}}, {{offset: 0, magic: []byte{0xFF, 0xF2}}}},
	".flac": {{{offset: 0, magic: []byte("fLaC")}}},
	".wav":  {{{offset: 0, magic: []byte("RIFF")}, {offset: 8, magic: []byte("WAVE")}}},
	".ogg":  {{{offset: 0, magic: []byte("OggS")}}},
	".m4a":  {{{offset: 4, magic: []byte("ftyp")}}},
	".aac":  {{{offset: 0, magic: []byte("ID3")}}, {{offset: 0, magic: []byte{0xFF, 0xF1}}}, {{offset: 0, magic: []byte{0xFF, 0xF9}}}},
	".aiff": {{{offset: 0, magic: []byte("FORM")}, {offset: 8, magic: []byte("AIFF")}}},

	// Video
	".mp4":  {{{offset: 4, magic: []byte("ftyp")}}},
	".mov":  {{{offset: 4, magic: []byte("ftyp")}}},
	".m4v":  {{{offset: 4, magic: []byte("ftyp")}}},
	".webm": {{{offset: 0, magic: []byte{0x1A, 0x45, 0xDF, 0xA3}}}},
	".mkv":  {{{offset: 0, magic: []byte{0x1A, 0x45, 0xDF, 0xA3}}}},
	".avi":  {{{offset: 0, magic: []byte("RIFF")}, {offset: 8, magic: []byte("AVI ")}}},
}

// Extension returns a file's lowercased extension, including the dot.
func Extension(fileName string) string {
	return strings.ToLower(filepath.Ext(fileName))
}

// ContentMatchesExtension reports whether the start of a file's content carries
// a marker consistent with its extension. An extension with no known marker
// returns true: for those the allowlist is the only check available, since
// there is nothing in the bytes to compare against.
func ContentMatchesExtension(fileName string, buffer []byte) bool {
	candidates, known := signatures[Extension(fileName)]
	if !known {
		return true
	}

	for _, candidate := range candidates {
		if matchesAll(candidate, buffer) {
			return true
		}
	}

	return false
}

func matchesAll(candidate candidate, buffer []byte) bool {
	for _, p := range candidate {
		if p.anywhere {
			if !bytes.Contains(buffer, p.magic) {
				return false
			}
			continue
		}
		if len(buffer) < p.offset+len(p.magic) {
			return false
		}
		if !bytes.Equal(buffer[p.offset:p.offset+len(p.magic)], p.magic) {
			return false
		}
	}

	return true
}
