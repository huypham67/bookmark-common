package requestutils

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	ErrFileRequired   = errors.New("file is required")
	ErrFileTooLarge   = errors.New("file too large")
	ErrFileTypeDenied = errors.New("file type not allowed")
)

// FileConstraint describes the policy a caller wants enforced on an upload.
// Every field is optional:
//   - MaxSizeBytes <= 0 disables the size limit (no body cap, no size check).
//   - AllowedExts empty accepts any extension.
type FileConstraint struct {
	FormKey      string
	MaxSizeBytes int64
	AllowedExts  []string // e.g. {".csv"}; matched case-insensitively
}

// FormFile extracts a single uploaded file and enforces a FileConstraint.
//
// It caps the request body with http.MaxBytesReader *before* reading, so an
// oversized upload is rejected without buffering it into memory — this is a
// security control (it prevents a client from exhausting memory by lying about
// Content-Length), which is why it lives here rather than being re-implemented
// per handler.
//
// On success, it returns the opened file; the caller is responsible for closing
// it. Failures are reported as the sentinel errors above (use errors.Is); the
// caller owns the HTTP status and message it maps them to. A non-sentinel error
// indicates an unexpected failure opening the file.
func FormFile(c *gin.Context, constraint FileConstraint) (multipart.File, *multipart.FileHeader, error) {
	if constraint.MaxSizeBytes > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, constraint.MaxSizeBytes)
	}

	fileHeader, err := c.FormFile(constraint.FormKey)
	if err != nil {
		// MaxBytesReader trips here when the body exceeds the limit.
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return nil, nil, ErrFileTooLarge
		}
		return nil, nil, ErrFileRequired
	}

	if constraint.MaxSizeBytes > 0 && fileHeader.Size > constraint.MaxSizeBytes {
		return nil, nil, ErrFileTooLarge
	}

	if !hasAllowedExt(fileHeader.Filename, constraint.AllowedExts) {
		return nil, nil, ErrFileTypeDenied
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, err
	}

	return file, fileHeader, nil
}

// hasAllowedExt reports whether filename's extension is allowed. An empty
// allow-list permits any extension.
func hasAllowedExt(filename string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}

	ext := filepath.Ext(filename)
	for _, a := range allowed {
		if strings.EqualFold(ext, a) {
			return true
		}
	}

	return false
}
