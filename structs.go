package main

type Client struct {
	Username        string
	UID             string
	GID             string
	OperatingSystem string
	Hostname        string
	PublicIP        string
	LocalIP         string
}

type Instructions struct {
	IsHeartbeat bool

	IsCommand bool

	IsKillswitch bool

	// IsDownload   bool
	// IsUpload     bool
	// FileName     string
	// FileContents []byte

	IsListFiles bool
	Path        string

	Message string
}

type Response struct {
	Successful bool
	Message    string

	// IsUpload     bool
	// FileName     string
	// FileContents []byte

	IsListFiles bool
	FileList
}
type FileList struct {
	File []Item
}

type Item struct {
	Name     string
	FullPath string
	IsFolder bool
}
