package rpc

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client is the aria2 RPC client interface
type Client interface {
	Protocol
	Close() error
}

type client struct {
	caller
	url   *url.URL
	token string
}

var (
	errInvalidParameter = errors.New("invalid parameter")
	errNotImplemented   = errors.New("not implemented")
	errConnTimeout      = errors.New("connect to aria2 daemon timeout")
)

// New returns an instance of Client
func New(ctx context.Context, uri string, token string, timeout time.Duration, notifier Notifier) (Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	var c caller
	switch u.Scheme {
	case "http", "https":
		c = newHTTPCaller(ctx, u, timeout, notifier)
	case "ws", "wss":
		c, err = newWebsocketCaller(ctx, u.String(), timeout, notifier)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errInvalidParameter
	}
	return &client{caller: c, url: u, token: token}, nil
}

// AddURI adds a new download by URI
func (c *client) AddURI(uri string, options ...interface{}) (gid string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, []string{uri})
	if options != nil {
		params = append(params, options...)
	}
	err = c.Call(aria2AddURI, params, &gid)
	return
}

// AddTorrent adds a BitTorrent download by uploading a ".torrent" file
func (c *client) AddTorrent(filename string, options ...interface{}) (gid string, err error) {
	co, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	file := base64.StdEncoding.EncodeToString(co)
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, file)
	if options != nil {
		params = append(params, options...)
	}
	err = c.Call(aria2AddTorrent, params, &gid)
	return
}

// AddMetalink adds a Metalink download
func (c *client) AddMetalink(filename string, options ...interface{}) (gid []string, err error) {
	co, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	file := base64.StdEncoding.EncodeToString(co)
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, file)
	if options != nil {
		params = append(params, options...)
	}
	err = c.Call(aria2AddMetalink, params, &gid)
	return
}

// Remove removes a download
func (c *client) Remove(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2Remove, params, &g)
	return
}

// ForceRemove forcefully removes a download
func (c *client) ForceRemove(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2ForceRemove, params, &g)
	return
}

// Pause pauses a download
func (c *client) Pause(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2Pause, params, &g)
	return
}

// PauseAll pauses all downloads
func (c *client) PauseAll() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2PauseAll, params, &ok)
	return
}

// ForcePause forcefully pauses a download
func (c *client) ForcePause(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2ForcePause, params, &g)
	return
}

// ForcePauseAll forcefully pauses all downloads
func (c *client) ForcePauseAll() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2ForcePauseAll, params, &ok)
	return
}

// Unpause unpauses a download
func (c *client) Unpause(gid string) (g string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2Unpause, params, &g)
	return
}

// UnpauseAll unpauses all downloads
func (c *client) UnpauseAll() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2UnpauseAll, params, &ok)
	return
}

// TellStatus returns the status of a download
func (c *client) TellStatus(gid string, keys ...string) (info StatusInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellStatus, params, &info)
	return
}

// GetURIs returns URIs used in a download
func (c *client) GetURIs(gid string) (infos []URIInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetURIs, params, &infos)
	return
}

// GetFiles returns file list of a download
func (c *client) GetFiles(gid string) (infos []FileInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetFiles, params, &infos)
	return
}

// GetPeers returns peers of a download (BitTorrent only)
func (c *client) GetPeers(gid string) (infos []PeerInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetPeers, params, &infos)
	return
}

// GetServers returns connected servers
func (c *client) GetServers(gid string) (infos []ServerInfo, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetServers, params, &infos)
	return
}

// TellActive returns active downloads
func (c *client) TellActive(keys ...string) (infos []StatusInfo, err error) {
	params := make([]interface{}, 0, 1)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellActive, params, &infos)
	return
}

// TellWaiting returns waiting downloads
func (c *client) TellWaiting(offset, num int, keys ...string) (infos []StatusInfo, err error) {
	params := make([]interface{}, 0, 3)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, offset)
	params = append(params, num)
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellWaiting, params, &infos)
	return
}

// TellStopped returns stopped downloads
func (c *client) TellStopped(offset, num int, keys ...string) (infos []StatusInfo, err error) {
	params := make([]interface{}, 0, 3)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, offset)
	params = append(params, num)
	if keys != nil {
		params = append(params, keys)
	}
	err = c.Call(aria2TellStopped, params, &infos)
	return
}

// ChangePosition changes position of a download in queue
func (c *client) ChangePosition(gid string, pos int, how string) (p int, err error) {
	params := make([]interface{}, 0, 3)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	params = append(params, pos)
	params = append(params, how)
	err = c.Call(aria2ChangePosition, params, &p)
	return
}

// ChangeURI changes URIs of a download
func (c *client) ChangeURI(gid string, fileindex int, delUris []string, addUris []string, position ...int) (p []int, err error) {
	params := make([]interface{}, 0, 5)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	params = append(params, fileindex)
	params = append(params, delUris)
	params = append(params, addUris)
	if position != nil {
		params = append(params, position[0])
	}
	err = c.Call(aria2ChangeURI, params, &p)
	return
}

// GetOption returns options of a download
func (c *client) GetOption(gid string) (m Option, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2GetOption, params, &m)
	return
}

// ChangeOption changes options of a download
func (c *client) ChangeOption(gid string, option Option) (ok string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	if option != nil {
		params = append(params, option)
	}
	err = c.Call(aria2ChangeOption, params, &ok)
	return
}

// GetGlobalOption returns global options
func (c *client) GetGlobalOption() (m Option, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetGlobalOption, params, &m)
	return
}

// ChangeGlobalOption changes global options
func (c *client) ChangeGlobalOption(options Option) (ok string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, options)
	err = c.Call(aria2ChangeGlobalOption, params, &ok)
	return
}

// GetGlobalStat returns global statistics
func (c *client) GetGlobalStat() (info GlobalStatInfo, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetGlobalStat, params, &info)
	return
}

// PurgeDownloadResult purges completed/error/removed downloads
func (c *client) PurgeDownloadResult() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2PurgeDownloadResult, params, &ok)
	return
}

// RemoveDownloadResult removes a completed/error/removed download
func (c *client) RemoveDownloadResult(gid string) (ok string, err error) {
	params := make([]interface{}, 0, 2)
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	params = append(params, gid)
	err = c.Call(aria2RemoveDownloadResult, params, &ok)
	return
}

// GetVersion returns aria2 version
func (c *client) GetVersion() (info VersionInfo, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetVersion, params, &info)
	return
}

// GetSessionInfo returns session information
func (c *client) GetSessionInfo() (info SessionInfo, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2GetSessionInfo, params, &info)
	return
}

// Shutdown shuts down aria2
func (c *client) Shutdown() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2Shutdown, params, &ok)
	return
}

// ForceShutdown forcefully shuts down aria2
func (c *client) ForceShutdown() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2ForceShutdown, params, &ok)
	return
}

// SaveSession saves current session
func (c *client) SaveSession() (ok string, err error) {
	params := []string{}
	if c.token != "" {
		params = append(params, "token:"+c.token)
	}
	err = c.Call(aria2SaveSession, params, &ok)
	return
}

// Multicall encapsulates multiple method calls
func (c *client) Multicall(methods []Method) (r []interface{}, err error) {
	if len(methods) == 0 {
		err = errInvalidParameter
		return
	}
	err = c.Call(aria2Multicall, methods, &r)
	return
}

// ListMethods returns all available RPC methods
func (c *client) ListMethods() (methods []string, err error) {
	err = c.Call(aria2ListMethods, []string{}, &methods)
	return
}

// caller interface for RPC calls
type caller interface {
	Call(method string, params, reply interface{}) (err error)
	Close() error
}

// httpCaller implements caller for HTTP
type httpCaller struct {
	uri    string
	c      *http.Client
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	once   sync.Once
}

func newHTTPCaller(ctx context.Context, u *url.URL, timeout time.Duration, notifier Notifier) *httpCaller {
	c := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1,
			MaxConnsPerHost:     1,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   3 * time.Second,
			ResponseHeaderTimeout: timeout,
		},
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	h := &httpCaller{uri: u.String(), c: c, cancel: cancel, wg: &wg}
	if notifier != nil {
		h.setNotifier(ctx, *u, notifier)
	}
	return h
}

func (h *httpCaller) Close() (err error) {
	h.once.Do(func() {
		h.cancel()
		h.wg.Wait()
	})
	return
}

func (h *httpCaller) setNotifier(ctx context.Context, u url.URL, notifier Notifier) (err error) {
	u.Scheme = "ws"
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return
	}
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		defer conn.Close()
		select {
		case <-ctx.Done():
			conn.SetWriteDeadline(time.Now().Add(time.Second))
			if err := conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Printf("sending websocket close message: %v", err)
			}
			return
		}
	}()
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		var request websocketResponse
		var err error
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err = conn.ReadJSON(&request); err != nil {
				select {
				case <-ctx.Done():
					return
				default:
				}
				log.Printf("conn.ReadJSON|err:%v", err.Error())
				return
			}
			switch request.Method {
			case "aria2.onDownloadStart":
				notifier.OnDownloadStart(request.Params)
			case "aria2.onDownloadPause":
				notifier.OnDownloadPause(request.Params)
			case "aria2.onDownloadStop":
				notifier.OnDownloadStop(request.Params)
			case "aria2.onDownloadComplete":
				notifier.OnDownloadComplete(request.Params)
			case "aria2.onDownloadError":
				notifier.OnDownloadError(request.Params)
			case "aria2.onBtDownloadComplete":
				notifier.OnBtDownloadComplete(request.Params)
			default:
				log.Printf("unexpected notification: %s", request.Method)
			}
		}
	}()
	return
}

func (h httpCaller) Call(method string, params, reply interface{}) (err error) {
	payload, err := EncodeClientRequest(method, params)
	if err != nil {
		return
	}
	r, err := h.c.Post(h.uri, "application/json", payload)
	if err != nil {
		return
	}
	err = DecodeClientResponse(r.Body, &reply)
	r.Body.Close()
	return
}

// websocketCaller implements caller for WebSocket
type websocketCaller struct {
	conn     *websocket.Conn
	sendChan chan *sendRequest
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
	once     sync.Once
	timeout  time.Duration
}

func newWebsocketCaller(ctx context.Context, uri string, timeout time.Duration, notifier Notifier) (*websocketCaller, error) {
	var header = http.Header{}
	conn, _, err := websocket.DefaultDialer.Dial(uri, header)
	if err != nil {
		return nil, err
	}

	sendChan := make(chan *sendRequest, 16)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	w := &websocketCaller{conn: conn, wg: &wg, cancel: cancel, sendChan: sendChan, timeout: timeout}
	processor := NewResponseProcessor()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			var resp websocketResponse
			if err := conn.ReadJSON(&resp); err != nil {
				select {
				case <-ctx.Done():
					return
				default:
				}
				log.Printf("conn.ReadJSON|err:%v", err.Error())
				return
			}
			if resp.Id == nil {
				if notifier != nil {
					switch resp.Method {
					case "aria2.onDownloadStart":
						notifier.OnDownloadStart(resp.Params)
					case "aria2.onDownloadPause":
						notifier.OnDownloadPause(resp.Params)
					case "aria2.onDownloadStop":
						notifier.OnDownloadStop(resp.Params)
					case "aria2.onDownloadComplete":
						notifier.OnDownloadComplete(resp.Params)
					case "aria2.onDownloadError":
						notifier.OnDownloadError(resp.Params)
					case "aria2.onBtDownloadComplete":
						notifier.OnBtDownloadComplete(resp.Params)
					default:
						log.Printf("unexpected notification: %s", resp.Method)
					}
				}
				continue
			}
			processor.Process(resp.ClientResponse)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		defer w.conn.Close()

		for {
			select {
			case <-ctx.Done():
				if err := w.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					log.Printf("sending websocket close message: %v", err)
				}
				return
			case req := <-sendChan:
				processor.Add(req.request.Id, func(resp ClientResponse) error {
					err := resp.decode(req.reply)
					req.cancel()
					return err
				})
				w.conn.SetWriteDeadline(time.Now().Add(timeout))
				w.conn.WriteJSON(req.request)
			}
		}
	}()

	return w, nil
}

func (w *websocketCaller) Close() (err error) {
	w.once.Do(func() {
		w.cancel()
		w.wg.Wait()
	})
	return
}

func (w websocketCaller) Call(method string, params, reply interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
	defer cancel()
	select {
	case w.sendChan <- &sendRequest{cancel: cancel, request: &clientRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
		Id:      reqid(),
	}, reply: reply}:

	default:
		return errors.New("sending channel blocking")
	}

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err == context.DeadlineExceeded {
			return err
		}
	}
	return
}

type sendRequest struct {
	cancel  context.CancelFunc
	request *clientRequest
	reply   interface{}
}
