package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadFile downloads a file from url to filepath.
func DownloadFile(url string, filepath string) error {
	var src io.ReadCloser
	if len(url) > 7 && url[:7] == "file://" {
		f, err := os.Open(url[7:])
		if err != nil {
			return err
		}
		src = f
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("bad status: %s", resp.Status)
		}
		src = resp.Body
	}
	defer src.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	n, err := io.Copy(out, src)
	if err != nil {
		return err
	}

	fmt.Printf("[Download] URL: %s -> %s (Size: %d bytes)\n", url, filepath, n)
	if n < 100 {
		return fmt.Errorf("downloaded file too small (%d bytes), likely failed", n)
	}
	return nil
}
