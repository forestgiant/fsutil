package fsutil

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func CheckIfCompressed(file io.ReadSeeker) bool {
	if file == nil {
		return false
	}

	// http://golang.org/pkg/net/http/#DetectContentType
	buff := make([]byte, 512)
	_, err := file.Read(buff)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// seek back to beginning of file
	file.Seek(0, 0)

	fileType := http.DetectContentType(buff)
	switch fileType {
	case "application/x-gzip", "application/zip":
		return true
	default:
		return false
	}
}

// outErr out param is set by defer function
func Unzip(src, dest string) (outErr error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			outErr = err
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	// closureErr out param is set by defer functions
	extractAndWriteFile := func(f *zip.File) (closureErr error) {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				closureErr = err
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					closureErr = err
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
