package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ScanRequest struct {
	Target string `json:"target"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
}

type PortDetail struct {
	Port       int    `json:"port"`
	Service    string `json:"service"`
	ResponseMs int64  `json:"response_ms"`
	Banner     string `json:"banner"`
}

type GeoIPInfo struct {
	Query      string  `json:"query"`
	Country    string  `json:"country"`
	RegionName string  `json:"regionName"`
	City       string  `json:"city"`
	ISP        string  `json:"isp"`
	Org        string  `json:"org"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
}

type ScanResult struct {
	Target string       `json:"target"`
	GeoIP  *GeoIPInfo   `json:"geoip,omitempty"`
	Ports  []PortDetail `json:"ports"`
}

var (
	lastScanResult *ScanResult
	lastScanMutex  sync.Mutex
)

var commonServices = map[int]string{
	21:   "FTP",
	22:   "SSH",
	23:   "Telnet",
	25:   "SMTP",
	53:   "DNS",
	80:   "HTTP",
	110:  "POP3",
	143:  "IMAP",
	443:  "HTTPS",
	3306: "MySQL",
	3389: "RDP",
	8080: "HTTP-alt",
}

func getGeoIP(target string) (*GeoIPInfo, error) {
	resp, err := http.Get("http://ip-api.com/json/" + target)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var info GeoIPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	geoip, _ := getGeoIP(req.Target)

	var wg sync.WaitGroup
	var mu sync.Mutex
	portDetails := []PortDetail{}

	for port := req.Start; port <= req.End; port++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			start := time.Now()
			address := fmt.Sprintf("%s:%d", req.Target, p)
			conn, err := net.DialTimeout("tcp", address, 300*time.Millisecond)
			responseMs := time.Since(start).Milliseconds()
			if err == nil {
				banner := ""
				conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
				buf := make([]byte, 128)
				n, _ := conn.Read(buf)
				if n > 0 {
					banner = string(buf[:n])
				}
				conn.Close()
				service := commonServices[p]
				mu.Lock()
				portDetails = append(portDetails, PortDetail{
					Port:       p,
					Service:    service,
					ResponseMs: responseMs,
					Banner:     banner,
				})
				mu.Unlock()
			}
		}(port)
	}
	wg.Wait()

	result := &ScanResult{
		Target: req.Target,
		GeoIP:  geoip,
		Ports:  portDetails,
	}

	lastScanMutex.Lock()
	lastScanResult = result
	lastScanMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func exportJSONHandler(w http.ResponseWriter, r *http.Request) {
	lastScanMutex.Lock()
	defer lastScanMutex.Unlock()
	if lastScanResult == nil {
		http.Error(w, "No scan result available", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=scan_result.json")
	json.NewEncoder(w).Encode(lastScanResult)
}

func exportCSVHandler(w http.ResponseWriter, r *http.Request) {
	lastScanMutex.Lock()
	defer lastScanMutex.Unlock()
	if lastScanResult == nil {
		http.Error(w, "No scan result available", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=scan_result.csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()
	writer.Write([]string{"Port", "Service", "ResponseMs", "Banner"})
	for _, pd := range lastScanResult.Ports {
		writer.Write([]string{
			strconv.Itoa(pd.Port),
			pd.Service,
			strconv.FormatInt(pd.ResponseMs, 10),
			pd.Banner,
		})
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/scan", scanHandler)
	http.HandleFunc("/export/json", exportJSONHandler)
	http.HandleFunc("/export/csv", exportCSVHandler)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
