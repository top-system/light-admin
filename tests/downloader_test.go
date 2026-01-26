package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/top-system/light-admin/pkg/downloader"
	"github.com/top-system/light-admin/pkg/downloader/aria2"
	"github.com/top-system/light-admin/pkg/downloader/aria2/rpc"
	"github.com/top-system/light-admin/pkg/downloader/qbittorrent"
)

// testLogger implements the Logger interface for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Info(format string, args ...interface{})    { l.t.Logf("[INFO] "+format, args...) }
func (l *testLogger) Debug(format string, args ...interface{})   { l.t.Logf("[DEBUG] "+format, args...) }
func (l *testLogger) Warning(format string, args ...interface{}) { l.t.Logf("[WARN] "+format, args...) }
func (l *testLogger) Error(format string, args ...interface{})   { l.t.Logf("[ERROR] "+format, args...) }

func TestDownloaderInterface(t *testing.T) {
	// Test that interfaces are correctly defined
	var _ downloader.Downloader = (*aria2.Client)(nil)
	var _ downloader.Downloader = (*qbittorrent.Client)(nil)
}

func TestTaskStatusMethods(t *testing.T) {
	t.Run("Progress calculation", func(t *testing.T) {
		status := &downloader.TaskStatus{
			Total:      1000,
			Downloaded: 500,
		}
		assert.Equal(t, 50.0, status.Progress())

		// Zero total
		status = &downloader.TaskStatus{
			Total:      0,
			Downloaded: 500,
		}
		assert.Equal(t, 0.0, status.Progress())
	})

	t.Run("IsComplete", func(t *testing.T) {
		status := &downloader.TaskStatus{State: downloader.StatusCompleted}
		assert.True(t, status.IsComplete())

		status = &downloader.TaskStatus{State: downloader.StatusDownloading}
		assert.False(t, status.IsComplete())
	})

	t.Run("IsError", func(t *testing.T) {
		status := &downloader.TaskStatus{State: downloader.StatusError}
		assert.True(t, status.IsError())

		status = &downloader.TaskStatus{State: downloader.StatusDownloading}
		assert.False(t, status.IsError())
	})

	t.Run("IsActive", func(t *testing.T) {
		status := &downloader.TaskStatus{State: downloader.StatusDownloading}
		assert.True(t, status.IsActive())

		status = &downloader.TaskStatus{State: downloader.StatusSeeding}
		assert.True(t, status.IsActive())

		status = &downloader.TaskStatus{State: downloader.StatusCompleted}
		assert.False(t, status.IsActive())
	})
}

func TestAria2ClientCreation(t *testing.T) {
	logger := &testLogger{t: t}

	t.Run("Create aria2 client with default settings", func(t *testing.T) {
		client := aria2.New(logger, &aria2.Settings{
			Server: "http://localhost:6800",
			Token:  "secret",
		})
		assert.NotNil(t, client)
	})

	t.Run("Create aria2 client with custom options", func(t *testing.T) {
		client := aria2.New(logger, &aria2.Settings{
			Server:   "http://localhost:6800",
			Token:    "secret",
			TempPath: "/tmp/downloads",
			Options: map[string]interface{}{
				"max-concurrent-downloads": 5,
			},
		})
		assert.NotNil(t, client)
	})

	t.Run("Create aria2 client with WebSocket URL", func(t *testing.T) {
		client := aria2.New(logger, &aria2.Settings{
			Server: "ws://localhost:6800",
			Token:  "secret",
		})
		assert.NotNil(t, client)
	})
}

func TestQBittorrentClientCreation(t *testing.T) {
	logger := &testLogger{t: t}

	t.Run("Create qBittorrent client with valid settings", func(t *testing.T) {
		client, err := qbittorrent.New(logger, &qbittorrent.Settings{
			Server:   "http://localhost:8080",
			User:     "admin",
			Password: "adminadmin",
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("Create qBittorrent client with invalid URL", func(t *testing.T) {
		_, err := qbittorrent.New(logger, &qbittorrent.Settings{
			Server: "://invalid-url",
		})
		assert.Error(t, err)
	})
}

func TestRPCJSONEncoding(t *testing.T) {
	t.Run("Encode client request", func(t *testing.T) {
		buf, err := rpc.EncodeClientRequest("aria2.getVersion", []interface{}{})
		require.NoError(t, err)
		assert.NotNil(t, buf)

		var req map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &req)
		require.NoError(t, err)
		assert.Equal(t, "2.0", req["jsonrpc"])
		assert.Equal(t, "aria2.getVersion", req["method"])
	})

	t.Run("Decode client response", func(t *testing.T) {
		response := `{"jsonrpc":"2.0","id":1,"result":"1.36.0"}`
		var version string
		err := rpc.DecodeClientResponse(strings.NewReader(response), &version)
		require.NoError(t, err)
		assert.Equal(t, "1.36.0", version)
	})

	t.Run("Decode error response", func(t *testing.T) {
		response := `{"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"Invalid request"}}`
		var result string
		err := rpc.DecodeClientResponse(strings.NewReader(response), &result)
		assert.Error(t, err)
	})
}

func TestRPCResponseProcessor(t *testing.T) {
	t.Run("Process response", func(t *testing.T) {
		proc := rpc.NewResponseProcessor()
		processed := false
		id := uint64(1)

		proc.Add(id, func(resp rpc.ClientResponse) error {
			processed = true
			return nil
		})

		err := proc.Process(rpc.ClientResponse{Id: &id})
		require.NoError(t, err)
		assert.True(t, processed)
	})

	t.Run("Process unregistered response", func(t *testing.T) {
		proc := rpc.NewResponseProcessor()
		id := uint64(999)

		err := proc.Process(rpc.ClientResponse{Id: &id})
		require.NoError(t, err)
	})
}

func TestMockAria2Server(t *testing.T) {
	// Create a mock aria2 RPC server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string        `json:"method"`
			ID     uint64        `json:"id"`
			Params []interface{} `json:"params"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var response interface{}
		switch req.Method {
		case "aria2.getVersion":
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]interface{}{
					"version":         "1.36.0",
					"enabledFeatures": []string{"BitTorrent", "Metalink"},
				},
			}
		case "aria2.addUri":
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result":  "2089b05ecca3d829",
			}
		case "aria2.tellStatus":
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]interface{}{
					"gid":             "2089b05ecca3d829",
					"status":          "active",
					"totalLength":     "1048576",
					"completedLength": "524288",
					"downloadSpeed":   "1024",
					"uploadSpeed":     "512",
					"files": []map[string]interface{}{
						{
							"index":           "1",
							"path":            "/tmp/test.file",
							"length":          "1048576",
							"completedLength": "524288",
							"selected":        "true",
						},
					},
				},
			}
		default:
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"error": map[string]interface{}{
					"code":    -32601,
					"message": "Method not found",
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	ctx := context.Background()

	t.Run("Test connection", func(t *testing.T) {
		client := aria2.New(&testLogger{t: t}, &aria2.Settings{
			Server: server.URL,
			Token:  "secret",
		})

		version, err := client.Test(ctx)
		require.NoError(t, err)
		assert.Equal(t, "1.36.0", version)
	})

	t.Run("Create task", func(t *testing.T) {
		client := aria2.New(&testLogger{t: t}, &aria2.Settings{
			Server:   server.URL,
			Token:    "secret",
			TempPath: t.TempDir(),
		})

		handle, err := client.CreateTask(ctx, "https://example.com/file.zip", nil)
		require.NoError(t, err)
		assert.NotEmpty(t, handle.ID)
	})

	t.Run("Get task info", func(t *testing.T) {
		client := aria2.New(&testLogger{t: t}, &aria2.Settings{
			Server: server.URL,
			Token:  "secret",
		})

		status, err := client.Info(ctx, &downloader.TaskHandle{ID: "2089b05ecca3d829"})
		require.NoError(t, err)
		assert.Equal(t, downloader.StatusDownloading, status.State)
		assert.Equal(t, int64(1048576), status.Total)
		assert.Equal(t, int64(524288), status.Downloaded)
		assert.Equal(t, 50.0, status.Progress())
	})
}

func TestMockQBittorrentServer(t *testing.T) {
	loggedIn := false

	// Create a mock qBittorrent server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v2/")

		// Check authentication for most endpoints
		if path != "auth/login" && !loggedIn {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		switch path {
		case "auth/login":
			loggedIn = true
			w.Write([]byte("Ok."))
		case "app/version":
			w.Write([]byte("v4.5.0"))
		case "torrents/add":
			w.Write([]byte("Ok."))
		case "torrents/info":
			response := []map[string]interface{}{
				{
					"hash":       "abc123",
					"name":       "Test Torrent",
					"size":       1048576,
					"completed":  524288,
					"dlspeed":    1024,
					"upspeed":    512,
					"uploaded":   256,
					"state":      "downloading",
					"save_path":  "/tmp/downloads",
					"progress":   0.5,
					"num_leechs": 5,
					"num_seeds":  10,
				},
			}
			json.NewEncoder(w).Encode(response)
		case "torrents/files":
			response := []map[string]interface{}{
				{
					"index":    0,
					"name":     "file1.txt",
					"size":     524288,
					"progress": 0.5,
					"priority": 1,
				},
				{
					"index":    1,
					"name":     "file2.txt",
					"size":     524288,
					"progress": 0.5,
					"priority": 0,
				},
			}
			json.NewEncoder(w).Encode(response)
		case "torrents/pieceStates":
			response := []int{2, 2, 2, 2, 0, 0, 0, 0}
			json.NewEncoder(w).Encode(response)
		case "torrents/delete":
			w.WriteHeader(http.StatusOK)
		case "torrents/deleteTags":
			w.WriteHeader(http.StatusOK)
		case "torrents/filePrio":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()

	t.Run("Test connection", func(t *testing.T) {
		client, err := qbittorrent.New(&testLogger{t: t}, &qbittorrent.Settings{
			Server:   server.URL,
			User:     "admin",
			Password: "adminadmin",
		})
		require.NoError(t, err)

		version, err := client.Test(ctx)
		require.NoError(t, err)
		assert.Equal(t, "v4.5.0", version)
	})

	t.Run("Create task", func(t *testing.T) {
		loggedIn = true // Ensure logged in state
		client, err := qbittorrent.New(&testLogger{t: t}, &qbittorrent.Settings{
			Server:   server.URL,
			User:     "admin",
			Password: "adminadmin",
			TempPath: t.TempDir(),
		})
		require.NoError(t, err)

		handle, err := client.CreateTask(ctx, "magnet:?xt=urn:btih:abc123", nil)
		require.NoError(t, err)
		assert.NotEmpty(t, handle.ID)
	})

	t.Run("Get task info", func(t *testing.T) {
		loggedIn = true
		client, err := qbittorrent.New(&testLogger{t: t}, &qbittorrent.Settings{
			Server:   server.URL,
			User:     "admin",
			Password: "adminadmin",
		})
		require.NoError(t, err)

		handle := &downloader.TaskHandle{ID: "test-id", Hash: "abc123"}
		status, err := client.Info(ctx, handle)
		require.NoError(t, err)
		assert.Equal(t, "Test Torrent", status.Name)
		assert.Equal(t, downloader.StatusDownloading, status.State)
		assert.Equal(t, int64(1048576), status.Total)
		assert.Len(t, status.Files, 2)
	})

	t.Run("Set files to download", func(t *testing.T) {
		loggedIn = true
		client, err := qbittorrent.New(&testLogger{t: t}, &qbittorrent.Settings{
			Server:   server.URL,
			User:     "admin",
			Password: "adminadmin",
		})
		require.NoError(t, err)

		handle := &downloader.TaskHandle{ID: "test-id", Hash: "abc123"}
		err = client.SetFilesToDownload(ctx, handle,
			&downloader.SetFileToDownloadArgs{Index: 0, Download: true},
			&downloader.SetFileToDownloadArgs{Index: 1, Download: false},
		)
		require.NoError(t, err)
	})

	t.Run("Cancel task", func(t *testing.T) {
		loggedIn = true
		client, err := qbittorrent.New(&testLogger{t: t}, &qbittorrent.Settings{
			Server:   server.URL,
			User:     "admin",
			Password: "adminadmin",
		})
		require.NoError(t, err)

		handle := &downloader.TaskHandle{ID: "test-id", Hash: "abc123"}
		err = client.Cancel(ctx, handle)
		require.NoError(t, err)
	})
}

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, downloader.Status("downloading"), downloader.StatusDownloading)
	assert.Equal(t, downloader.Status("seeding"), downloader.StatusSeeding)
	assert.Equal(t, downloader.Status("completed"), downloader.StatusCompleted)
	assert.Equal(t, downloader.Status("error"), downloader.StatusError)
	assert.Equal(t, downloader.Status("unknown"), downloader.StatusUnknown)
}

func TestErrTaskNotFound(t *testing.T) {
	assert.NotNil(t, downloader.ErrTaskNotFound)
	assert.Equal(t, "task not found", downloader.ErrTaskNotFound.Error())
}
