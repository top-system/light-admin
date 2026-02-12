package errors

import "net/http"

var (
	DownloadQueueNotEnabled    = New("task queue is not enabled")
	DownloadNoDownloaderConfig = New("no downloader configured")
	DownloadDownloaderNotFound = New("downloader not found")
)

func init() {
	RegisterHTTPStatus(DownloadDownloaderNotFound, http.StatusNotFound)
	RegisterHTTPStatus(DownloadQueueNotEnabled, http.StatusServiceUnavailable)
	RegisterHTTPStatus(DownloadNoDownloaderConfig, http.StatusServiceUnavailable)
}
