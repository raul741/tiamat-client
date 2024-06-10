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
	IsHeartbeat  bool
	IsCommand    bool
	IsKillswitch bool
	Message      string
}

type Response struct {
	Successful bool
	Message    string
}
