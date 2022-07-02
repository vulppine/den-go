package files

import (
	"bytes"
	"den/routing"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler transmits a single file over, from a specific directory on
// the current host's filesystem. If the handler fails to grab a file,
// it will return a ResponseInfo of type text, with the raw error
// in question. Otherwise, it will return the data file's reader.
type FileHandler struct {
	basePath string
}

func NewFileHandler(path string) *FileHandler {
	handler := new(FileHandler)
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	handler.basePath = path

	return handler
}

func (f *FileHandler) HandleRequest(req *routing.RequestInfo) (*routing.ResponseInfo, error) {
	path := filepath.FromSlash(strings.Join(req.Path, "/"))

	if req.Method() != http.MethodGet {
		err := newFileHandlerError(invalidMethod, path, nil)
		text := bytes.NewBufferString(err.Error())

		resp := routing.CreateResponseInfo(http.StatusBadRequest, http.Header{}, routing.Text, req.RequestEndpoint(), text)

		return &resp, nil
	}

	file, err := f.getFileAtPath(path)

	if err != nil {
		e := err.(fileHandlerError)
		var code int

		switch e.code {
		case notAllowed:
			code = http.StatusForbidden
		case accessError:
			switch {
			case errors.Is(err, fs.ErrNotExist):
				code = http.StatusNotFound
			case errors.Is(err, fs.ErrPermission):
				code = http.StatusForbidden
			default:
				// something really odd happened, so
				// it might be a server-side thing
				code = http.StatusServiceUnavailable
			}
		}

		text := bytes.NewBufferString(e.Error())

		resp := routing.CreateResponseInfo(code, http.Header{}, routing.Text, req.RequestEndpoint(), text)

		return &resp, nil
	}

	resp := routing.CreateResponseInfo(http.StatusOK, http.Header{}, routing.Data, req.RequestEndpoint(), file)

	return &resp, nil
}

func (f *FileHandler) getFileAtPath(path string) (*os.File, error) {
	// only absolute file paths allowed here, buddy
	if !filepath.IsAbs(path) {
		return nil, newFileHandlerError(notAllowed, path, errors.New("relative file pathing not allowed"))
	}

	fullPath := filepath.Join(f.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, newFileHandlerError(accessError, path, err)
	}

	return file, nil
}
