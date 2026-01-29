package storage

import (
	"github.com/gabriel-vasile/mimetype"
	"bytes"
	"io"
)

func detectMimeType(r io.Reader) (string, io.Reader, error) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	mime, err := mimetype.DetectReader(tee)
	if err != nil {
		return "", nil, err
	}

	// Reconstruct reader: buffered bytes + remaining stream
	newReader := io.MultiReader(&buf, r)

	return mime.String(), newReader, nil
}