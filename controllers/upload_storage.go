package controllers

import (
	"compress/gzip"
	"io"
	"os"
)

func storeUploadedDocument(source io.ReadSeeker, basePath, extension string, originalSize int64) (string, error) {
	compressedPath := basePath + extension + ".gz"
	if err := writeUploadedDocument(compressedPath, source, true); err != nil {
		return "", err
	}

	compressedInfo, err := os.Stat(compressedPath)
	if err != nil {
		_ = os.Remove(compressedPath)
		return "", err
	}
	if compressedInfo.Size() < originalSize {
		return compressedPath, nil
	}
	if err := os.Remove(compressedPath); err != nil {
		return "", err
	}

	rawPath := basePath + extension
	if err := writeUploadedDocument(rawPath, source, false); err != nil {
		return "", err
	}
	return rawPath, nil
}

func writeUploadedDocument(path string, source io.ReadSeeker, compress bool) error {
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return err
	}

	destination, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	var copyErr error
	if compress {
		writer := gzip.NewWriter(destination)
		_, copyErr = io.Copy(writer, source)
		if closeErr := writer.Close(); copyErr == nil {
			copyErr = closeErr
		}
	} else {
		_, copyErr = io.Copy(destination, source)
	}
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(path)
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}
	return nil
}
