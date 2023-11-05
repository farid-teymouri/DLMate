package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// downloadFile downloads a file from the specified URL and saves it to the given destination path.
// It handles creating directories, making an HTTP GET request, and copying the file content.
//
// url: The URL of the file to download.
// destPath: The path where the downloaded file will be saved.
//
// Returns an error if any step of the download process encounters an issue.
func downloadFile(url, destPath string) error {
    // Create the directory structure for the destination path.
    dir := filepath.Dir(destPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Make an HTTP GET request to the specified URL.
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Create a file at the destination path and prepare to write the downloaded content.
    out, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Copy the content from the HTTP response to the destination file.
    _, err = io.Copy(out, resp.Body)

    // Return any encountered errors during the process, if any.
    return err
}