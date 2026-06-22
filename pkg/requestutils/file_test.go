package requestutils

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fileUpload builds a multipart request body carrying one file field.
func fileUpload(t *testing.T, field, filename, content string) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return &body, writer.FormDataContentType()
}

func newTestContext(t *testing.T, body *bytes.Buffer, contentType string) *gin.Context {
	t.Helper()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	ctx.Request = req

	return ctx
}

func TestFormFile(t *testing.T) {
	t.Parallel()

	const oneMiB = 1 << 20

	testCases := []struct {
		name       string
		constraint FileConstraint
		buildBody  func(*testing.T) (*bytes.Buffer, string)
		verify     func(*testing.T, multipart.File, *multipart.FileHeader, error)
	}{
		{
			name:       "returns the opened file when constraints pass",
			constraint: FileConstraint{FormKey: "file", MaxSizeBytes: oneMiB, AllowedExts: []string{".csv"}},
			buildBody: func(t *testing.T) (*bytes.Buffer, string) {
				return fileUpload(t, "file", "data.csv", "description,url\n")
			},
			verify: func(t *testing.T, file multipart.File, header *multipart.FileHeader, err error) {
				require.NoError(t, err)
				require.NotNil(t, file)
				defer file.Close()

				assert.Equal(t, "data.csv", header.Filename)
				content, rerr := io.ReadAll(file)
				require.NoError(t, rerr)
				assert.Equal(t, "description,url\n", string(content))
			},
		},
		{
			name:       "ErrFileRequired when the form field is missing",
			constraint: FileConstraint{FormKey: "file", MaxSizeBytes: oneMiB},
			buildBody: func(t *testing.T) (*bytes.Buffer, string) {
				// A file under a different key.
				return fileUpload(t, "other", "data.csv", "x")
			},
			verify: func(t *testing.T, _ multipart.File, _ *multipart.FileHeader, err error) {
				assert.ErrorIs(t, err, ErrFileRequired)
			},
		},
		{
			name:       "ErrFileTypeDenied when extension is not allowed",
			constraint: FileConstraint{FormKey: "file", MaxSizeBytes: oneMiB, AllowedExts: []string{".csv"}},
			buildBody: func(t *testing.T) (*bytes.Buffer, string) {
				return fileUpload(t, "file", "data.txt", "x")
			},
			verify: func(t *testing.T, _ multipart.File, _ *multipart.FileHeader, err error) {
				assert.ErrorIs(t, err, ErrFileTypeDenied)
			},
		},
		{
			name:       "ErrFileTooLarge when body exceeds the limit",
			constraint: FileConstraint{FormKey: "file", MaxSizeBytes: 64, AllowedExts: []string{".csv"}},
			buildBody: func(t *testing.T) (*bytes.Buffer, string) {
				return fileUpload(t, "file", "data.csv", strings.Repeat("a", 256))
			},
			verify: func(t *testing.T, _ multipart.File, _ *multipart.FileHeader, err error) {
				assert.ErrorIs(t, err, ErrFileTooLarge)
			},
		},
		{
			name:       "empty allow-list accepts any extension",
			constraint: FileConstraint{FormKey: "file", MaxSizeBytes: oneMiB},
			buildBody: func(t *testing.T) (*bytes.Buffer, string) {
				return fileUpload(t, "file", "data.txt", "x")
			},
			verify: func(t *testing.T, file multipart.File, _ *multipart.FileHeader, err error) {
				require.NoError(t, err)
				require.NotNil(t, file)
				_ = file.Close()
			},
		},
		{
			name:       "zero size limit disables the size check",
			constraint: FileConstraint{FormKey: "file", AllowedExts: []string{".csv"}},
			buildBody: func(t *testing.T) (*bytes.Buffer, string) {
				return fileUpload(t, "file", "data.csv", strings.Repeat("a", 4096))
			},
			verify: func(t *testing.T, file multipart.File, _ *multipart.FileHeader, err error) {
				require.NoError(t, err)
				require.NotNil(t, file)
				_ = file.Close()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, contentType := tc.buildBody(t)
			ctx := newTestContext(t, body, contentType)

			file, header, err := FormFile(ctx, tc.constraint)

			tc.verify(t, file, header, err)
		})
	}
}
