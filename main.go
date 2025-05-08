package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type ScanRequest struct {
	Target string `json:"target"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
}

type ScanResult struct {
	OpenPorts []int `json:"open_ports"`
}

func scanPort(target string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	json.NewDecoder(r.Body).Decode(&req)
	openPorts := []int{}
	for port := req.Start; port <= req.End; port++ {
		if scanPort(req.Target, port, 500*time.Millisecond) {
			openPorts = append(openPorts, port)
		}
	}
	res := ScanResult{OpenPorts: openPorts}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/scan", scanHandler)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
