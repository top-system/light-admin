package rpc

// StatusInfo represents response of aria2.tellStatus
type StatusInfo struct {
	Gid             string         `json:"gid"`             // GID of the download
	Status          string         `json:"status"`          // active, waiting, paused, error, complete, removed
	TotalLength     string         `json:"totalLength"`     // Total length of the download in bytes
	CompletedLength string         `json:"completedLength"` // Completed length of the download in bytes
	UploadLength    string         `json:"uploadLength"`    // Uploaded length of the download in bytes
	BitField        string         `json:"bitfield"`        // Hexadecimal representation of the download progress
	DownloadSpeed   string         `json:"downloadSpeed"`   // Download speed in bytes/sec
	UploadSpeed     string         `json:"uploadSpeed"`     // Upload speed in bytes/sec
	InfoHash        string         `json:"infoHash"`        // InfoHash (BitTorrent only)
	NumSeeders      string         `json:"numSeeders"`      // Number of seeders (BitTorrent only)
	Seeder          string         `json:"seeder"`          // true if local endpoint is a seeder (BitTorrent only)
	PieceLength     string         `json:"pieceLength"`     // Piece length in bytes
	NumPieces       string         `json:"numPieces"`       // Number of pieces
	Connections     string         `json:"connections"`     // Number of peers/servers connected
	ErrorCode       string         `json:"errorCode"`       // Error code if any
	ErrorMessage    string         `json:"errorMessage"`    // Error message
	FollowedBy      []string       `json:"followedBy"`      // List of GIDs generated as result of this download
	BelongsTo       string         `json:"belongsTo"`       // GID of parent download
	Dir             string         `json:"dir"`             // Directory to save files
	Files           []FileInfo     `json:"files"`           // List of files
	BitTorrent      BitTorrentInfo `json:"bittorrent"`      // BitTorrent info
}

// URIInfo represents an element of response of aria2.getUris
type URIInfo struct {
	URI    string `json:"uri"`    // URI
	Status string `json:"status"` // 'used' or 'waiting'
}

// FileInfo represents an element of response of aria2.getFiles
type FileInfo struct {
	Index           string    `json:"index"`           // Index of the file, starting at 1
	Path            string    `json:"path"`            // File path
	Length          string    `json:"length"`          // File size in bytes
	CompletedLength string    `json:"completedLength"` // Completed length in bytes
	Selected        string    `json:"selected"`        // true if selected by --select-file option
	URIs            []URIInfo `json:"uris"`            // List of URIs for this file
}

// PeerInfo represents an element of response of aria2.getPeers
type PeerInfo struct {
	PeerId        string `json:"peerId"`        // Percent-encoded peer ID
	IP            string `json:"ip"`            // IP address
	Port          string `json:"port"`          // Port number
	BitField      string `json:"bitfield"`      // Hexadecimal representation of download progress
	AmChoking     string `json:"amChoking"`     // true if aria2 is choking the peer
	PeerChoking   string `json:"peerChoking"`   // true if the peer is choking aria2
	DownloadSpeed string `json:"downloadSpeed"` // Download speed from peer
	UploadSpeed   string `json:"uploadSpeed"`   // Upload speed to peer
	Seeder        string `json:"seeder"`        // true if peer is a seeder
}

// ServerInfo represents an element of response of aria2.getServers
type ServerInfo struct {
	Index   string `json:"index"` // Index of the file
	Servers []struct {
		URI           string `json:"uri"`           // Original URI
		CurrentURI    string `json:"currentUri"`    // Current URI (may differ due to redirect)
		DownloadSpeed string `json:"downloadSpeed"` // Download speed
	} `json:"servers"` // List of servers
}

// GlobalStatInfo represents response of aria2.getGlobalStat
type GlobalStatInfo struct {
	DownloadSpeed   string `json:"downloadSpeed"`   // Overall download speed
	UploadSpeed     string `json:"uploadSpeed"`     // Overall upload speed
	NumActive       string `json:"numActive"`       // Number of active downloads
	NumWaiting      string `json:"numWaiting"`      // Number of waiting downloads
	NumStopped      string `json:"numStopped"`      // Number of stopped downloads (capped)
	NumStoppedTotal string `json:"numStoppedTotal"` // Number of stopped downloads (total)
}

// VersionInfo represents response of aria2.getVersion
type VersionInfo struct {
	Version  string   `json:"version"`         // Version number
	Features []string `json:"enabledFeatures"` // List of enabled features
}

// SessionInfo represents response of aria2.getSessionInfo
type SessionInfo struct {
	Id string `json:"sessionId"` // Session ID
}

// Method is an element of parameters used in system.multicall
type Method struct {
	Name   string        `json:"methodName"` // Method name to call
	Params []interface{} `json:"params"`     // Parameters
}

// BitTorrentInfo contains BitTorrent specific information
type BitTorrentInfo struct {
	AnnounceList [][]string `json:"announceList"` // List of announce URIs
	Comment      string     `json:"comment"`      // Torrent comment
	CreationDate int64      `json:"creationDate"` // Creation time (epoch)
	Mode         string     `json:"mode"`         // File mode: single or multi
	Info         struct {
		Name string `json:"name"` // Name from info dictionary
	} `json:"info"` // Info dictionary data
}

// Option is a container for specifying Call parameters
type Option map[string]interface{}
