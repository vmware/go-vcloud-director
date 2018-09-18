package util

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Unpack(tarFile string) ([]string, error) {

	var filePaths []string

	reader, err := os.Open(tarFile)
	if err != nil {
		return filePaths, err
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)

	// creating dst
	dir, _ := filepath.Split(tarFile)
	var dst = dir + "/temp"
	if _, err := os.Stat(dst); err != nil {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return filePaths, err
		}
	}

	var expectedFileSize int64 = -1

	for {
		header, err := tarReader.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return filePaths, nil

			// return any other error
		case err != nil:
			return filePaths, err

			// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue

		case header != nil:
			expectedFileSize = header.Size
		}

		// the target location where the dir/newFile should be created
		target := filepath.Join(dst, header.Name)
		GovcdLogger.Printf("[TRACE] extracting newFile: %s \n", target)

		// check the newFile type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return filePaths, err
				}
			}

		case tar.TypeSymlink:
			if header.Linkname != "" {
				err := os.Symlink(header.Linkname, target)
				if err != nil {
					return filePaths, err
				}
			} else {
				return filePaths, errors.New("File %s is a symlink, but no link information was provided\n")
			}

			// if it's a newFile create it
		case tar.TypeReg:
			newFile, err := os.OpenFile(sanitizedName(target), os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return filePaths, err
			}

			// copy over contents
			if _, err := io.Copy(newFile, tarReader); err != nil {
				return filePaths, err
			}

			filePaths = append(filePaths, newFile.Name())

			if err := isExtractedFileValid(newFile, expectedFileSize); err != nil {
				newFile.Close()
				return filePaths, err
			}

			// manually close here after each newFile operation; defering would cause each newFile close
			// to wait until all operations have completed.
			newFile.Close()
		}
	}
}

func isExtractedFileValid(file *os.File, expectedFileSize int64) error {
	if fInfo, err := file.Stat(); err == nil {
		GovcdLogger.Printf("[TRACE] isExtractedFileValid: created file size %#v, size from header %#v.\n", fInfo.Size(), expectedFileSize)
		if fInfo.Size() != expectedFileSize && expectedFileSize != -1 {
			return errors.New("extracted file didn't match defined file size")
		}
	}
	return nil
}

func sanitizedName(filename string) string {
	if len(filename) > 1 && filename[1] == ':' {
		filename = filename[2:]
	}
	filename = strings.Replace(filename, "\\/.", "\\/", -1)
	filename = strings.Replace(filename, "../", "", -1)
	return strings.Replace(filename, "..\\", "", -1)
}
