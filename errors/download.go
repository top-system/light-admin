package errors

var (
	DownloadQueueNotEnabled    = New("task queue is not enabled")
	DownloadNoDownloaderConfig = New("no downloader configured")
	DownloadDownloaderNotFound = New("downloader not found")
)
