package file_util

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/util/encrypt"
	"io"
	"net/http"
	"os"
)

func DownloadImgFromUrl(url string) (*os.File, string, error) {
	// Create the file
	var filePath string
	path := GetFileRoot()
	filePath = fmt.Sprintf("%s/%d.png", path, encrypt.HashStr(url))
	file, err := os.Create(filePath)
	if err != nil {
		return file, filePath, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Download the image
	resp, err := http.Get(url)
	if err != nil {
		return file, filePath, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Write the image data to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return file, filePath, err
	}
	return file, filePath, err
}
