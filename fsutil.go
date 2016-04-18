package fsutil

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// CheckIfCompressed asserts whether the designated file is compressed
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

// Unzip extracts the contents at the designated src path and writes them to the destination path
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

// FileExists asserts whether provided path results in a valid file on the system
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

//CopyFile copies the file at the source path to the provided destination.
func CopyFile(source, destination string) error {
	//Validate the source and destination paths
	if len(source) == 0 {
		return errors.New("You must provide a source file path.")
	}

	if len(destination) == 0 {
		return errors.New("You must provide a destination file path.")
	}

	//Verify the source path refers to a regular file
	sourceFileInfo, err := os.Lstat(source)
	if err != nil {
		return err
	}

	//Handle regular files differently than symbolic links and other non-regular files.
	if sourceFileInfo.Mode().IsRegular() {
		//open the source file
		sourceFile, err := os.Open(source)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		//create the destinatin file
		destinationFile, err := os.Create(destination)
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		//copy the source file contents to the destination file
		if _, err = io.Copy(destinationFile, sourceFile); err != nil {
			return err
		}

		//replicate the source file mode for the destination file
		if err := os.Chmod(destination, sourceFileInfo.Mode()); err != nil {
			return err
		}
	} else if sourceFileInfo.Mode()&os.ModeSymlink != 0 {
		linkDestinaton, err := os.Readlink(source)
		if err != nil {
			return errors.New("Unable to read symlink. " + err.Error())
		}

		if err := os.Symlink(linkDestinaton, destination); err != nil {
			return errors.New("Unable to replicate symlink. " + err.Error())
		}
	} else {
		return errors.New("Unable to use io.Copy on file with mode " + string(sourceFileInfo.Mode()))
	}

	return nil
}

//CopyDirectory copies the directory at the source path to the provided destination, with the option of recursively copying subdirectories.
func CopyDirectory(source string, destination string, recursive bool) error {
	if len(source) == 0 || len(destination) == 0 {
		return errors.New("File paths must not be empty.")
	}

	//get properties of the source directory
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	//create the destination directory
	err = os.MkdirAll(destination, sourceInfo.Mode())
	if err != nil {
		return err
	}

	sourceDirectory, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceDirectory.Close()

	objects, err := sourceDirectory.Readdir(-1)
	if err != nil {
		return err
	}

	for _, object := range objects {
		if object.Name() == ".Trashes" {
			continue
		}

		sourceObjectName := source + "/" + object.Name()
		destObjectName := destination + "/" + object.Name()

		if object.IsDir() {
			//create sub-directories
			err = CopyDirectory(sourceObjectName, destObjectName, true)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(sourceObjectName, destObjectName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsEmpty test if a directory is empty
func IsEmpty(path string) (bool, error) {
	d, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer d.Close()

	_, err = d.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}

// RemoveDirContent removes all the files in a directory
// but keeps the directory
func RemoveDirContent(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}

	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}
