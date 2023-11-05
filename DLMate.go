package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"golang.org/x/sync/semaphore"
)

// maxConcurrentDownloads defines the maximum number of concurrent downloads allowed.
const maxConcurrentDownloads = 5

// main is the entry point of the download application.
func main() {
	// Initialize an empty list to store the user-provided URLs.
	var urls []string

	// Prompt the user for input and collect a list of URLs to download.
	fmt.Println("Enter the URLs you want to download (one per line). Enter an empty line to start downloading:")
	for {
		var url string
		_, err := fmt.Scanln(&url)
		if err != nil {
			break
		}
		urls = append(urls, url)
	}

	// If no URLs are provided, display a message and exit.
	if len(urls) == 0 {
		fmt.Println("No URLs provided. Exiting.")
		return
	}

	// Create a wait group to manage the download workers and a progress bar container.
	var wg sync.WaitGroup
	p := mpb.New()

	// Initialize a semaphore to control the maximum concurrent downloads.
	sem := semaphore.NewWeighted(maxConcurrentDownloads)

	// Iterate through the list of URLs and initiate download workers.
	for index, url := range urls {
		wg.Add(1)
		sem.Acquire(context.Background(), 1) // Acquire a semaphore slot

		go func(url string, index int) {
			defer wg.Done()
			defer sem.Release(1) // Release the semaphore slot

			// Extract the file name from the URL.
			fileName := filepath.Base(url)
			destPath := filepath.Join("download", fileName)

			// Call the download function with progress tracking.
			err := downloadFileWithProgress(url, destPath, p)
			if err != nil {
				fmt.Printf("Error downloading file from %s: %v\n", url, err)
			}
		}(url, index)
	}

	// Wait for all download workers to complete and for progress bars to finish.
	wg.Wait()
	p.Wait()
}

// downloadFileWithProgress downloads a file from a given URL to the specified destination path
// with progress tracking using mpb (a progress bar library).
// It also sets a custom User-Agent for the HTTP request.
//
// url: The URL of the file to download.
// destPath: The destination path to save the downloaded file.
// p: An mpb.Progress instance for tracking the download progress.
//
// Returns an error if any part of the download process fails.
func downloadFileWithProgress(url, destPath string, p *mpb.Progress) error {
	// Create the directory to store the downloaded file.
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Set a custom User-Agent for the HTTP request.
	customUserAgent := "asandev.com"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", customUserAgent)

	// Create an HTTP client and send the request to download the file.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the local file to save the downloaded content.
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Extract the filename from the destination path.
	filename := filepath.Base(destPath)

	// Create a progress bar for tracking the download.
	bar := p.AddBar(
		resp.ContentLength,
		mpb.PrependDecorators(
			decor.Name(filename+" "), // Display the filename in the progress bar.
			decor.CountersKibiByte("% .2f / % .2f"), // Show download progress in KiB.
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_MMSS, 60), // Display estimated time of arrival.
			decor.Name(" ] "), // Custom separator.
			decor.AverageSpeed(decor.UnitKB, "% .2f"), // Display average download speed.
		),
	)

	// Create a proxy reader for the progress bar and copy the response to the local file.
	proxyReader := bar.ProxyReader(resp.Body)
	defer proxyReader.Close()

	_, err = io.Copy(out, proxyReader)
	return err
}