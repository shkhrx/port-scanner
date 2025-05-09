// Package main implements a concurrent TCP port scanner with GeoIP lookup capabilities
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

// ScanRequest represents the structure of the incoming scan request from the frontend
type ScanRequest struct {
	Target string `json:"target"` // IP address or domain name to scan
	Start  int    `json:"start"`  // Starting port number
	End    int    `json:"end"`    // Ending port number
}

// PortDetail holds information about a single scanned port
type PortDetail struct {
	Port       int    `json:"port"`        // Port number
	Service    string `json:"service"`     // Common service name (e.g., HTTP, SSH)
	ResponseMs int64  `json:"response_ms"` // Response time in milliseconds
	Banner     string `json:"banner"`      // Banner information received from the port
}

// GeoIPInfo contains geographical and network information about the target
type GeoIPInfo struct {
	Query      string  `json:"query"`      // IP address of the target
	Country    string  `json:"country"`    // Country name
	RegionName string  `json:"regionName"` // Region/State name
	City       string  `json:"city"`       // City name
	ISP        string  `json:"isp"`        // Internet Service Provider
	Org        string  `json:"org"`        // Organization name
	Lat        float64 `json:"lat"`        // Latitude coordinate
	Lon        float64 `json:"lon"`        // Longitude coordinate
}

// ScanResult represents the complete result of a port scan
type ScanResult struct {
	Target string       `json:"target"`          // Target IP/domain
	GeoIP  *GeoIPInfo   `json:"geoip,omitempty"` // Geographical information
	Ports  []PortDetail `json:"ports"`           // List of open ports and their details
}

// Global variables for storing the last scan result
var (
	lastScanResult *ScanResult // Stores the most recent scan result
	lastScanMutex  sync.Mutex  // Mutex to protect concurrent access to lastScanResult
)

// commonServices maps well-known port numbers to their service names
var commonServices = map[int]string{
	21:   "FTP",      // File Transfer Protocol
	22:   "SSH",      // Secure Shell
	23:   "Telnet",   // Telnet
	25:   "SMTP",     // Simple Mail Transfer Protocol
	53:   "DNS",      // Domain Name System
	80:   "HTTP",     // Hypertext Transfer Protocol
	110:  "POP3",     // Post Office Protocol
	143:  "IMAP",     // Internet Message Access Protocol
	443:  "HTTPS",    // HTTP Secure
	3306: "MySQL",    // MySQL Database
	3389: "RDP",      // Remote Desktop Protocol
	8080: "HTTP-alt", // Alternative HTTP port
}

// getGeoIP retrieves geographical information for the target using ip-api.com
func getGeoIP(target string) (*GeoIPInfo, error) {
	// Make HTTP request to ip-api.com
	resp, err := http.Get("http://ip-api.com/json/" + target)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode JSON response into GeoIPInfo struct
	var info GeoIPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// scanHandler processes port scan requests from the frontend
func scanHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the incoming JSON request
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get geographical information for the target
	geoip, _ := getGeoIP(req.Target)

	// Initialize synchronization primitives
	var wg sync.WaitGroup
	var mu sync.Mutex
	portDetails := []PortDetail{}

	// Create a semaphore to limit concurrent connections
	sem := make(chan struct{}, 100) // Maximum 100 concurrent connections

	// Scan each port in the specified range
	for port := req.Start; port <= req.End; port++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }() // Release semaphore

			// Attempt to connect to the port
			start := time.Now()
			address := fmt.Sprintf("%s:%d", req.Target, p)
			conn, err := net.DialTimeout("tcp", address, 1*time.Second)
			responseMs := time.Since(start).Milliseconds()

			// If connection successful, try to read banner
			if err == nil {
				banner := ""
				conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
				buf := make([]byte, 256)
				n, _ := conn.Read(buf)
				if n > 0 {
					banner = string(buf[:n])
				}
				conn.Close()

				// Get service name from common services map
				service := commonServices[p]

				// Add port details to results
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

	// Create a channel to handle timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		// Scan completed successfully
	case <-time.After(30 * time.Second):
		// Scan timed out after 30 seconds
	}

	// Create and store the scan result
	result := &ScanResult{
		Target: req.Target,
		GeoIP:  geoip,
		Ports:  portDetails,
	}

	// Store the result for export
	lastScanMutex.Lock()
	lastScanResult = result
	lastScanMutex.Unlock()

	// Send the result back to the frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// exportJSONHandler exports the last scan result as a JSON file
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

// exportCSVHandler exports the last scan result as a CSV file
func exportCSVHandler(w http.ResponseWriter, r *http.Request) {
	lastScanMutex.Lock()
	defer lastScanMutex.Unlock()
	if lastScanResult == nil {
		http.Error(w, "No scan result available", http.StatusNotFound)
		return
	}

	// Debug print to verify GeoIP data
	fmt.Printf("Exporting CSV for target: %s\n", lastScanResult.Target)
	if lastScanResult.GeoIP != nil {
		fmt.Printf("GeoIP data available: Country=%s, City=%s\n",
			lastScanResult.GeoIP.Country,
			lastScanResult.GeoIP.City)
	} else {
		fmt.Println("No GeoIP data available")
	}

	// Set up CSV writer
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=scan_result.csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	headers := []string{
		"Target",
		"Country",
		"Region",
		"City",
		"ISP",
		"Organization",
		"Port",
		"Service",
		"ResponseMs",
		"Banner",
	}
	writer.Write(headers)

	// Write data rows
	for _, pd := range lastScanResult.Ports {
		row := make([]string, 0, 10)

		// Add target
		row = append(row, lastScanResult.Target)

		// Add GeoIP info if available
		if lastScanResult.GeoIP != nil {
			row = append(row,
				lastScanResult.GeoIP.Country,
				lastScanResult.GeoIP.RegionName,
				lastScanResult.GeoIP.City,
				lastScanResult.GeoIP.ISP,
				lastScanResult.GeoIP.Org,
			)
		} else {
			// Add empty strings for missing GeoIP data
			row = append(row, "", "", "", "", "")
		}

		// Add port details
		row = append(row,
			strconv.Itoa(pd.Port),
			pd.Service,
			strconv.FormatInt(pd.ResponseMs, 10),
			pd.Banner,
		)

		writer.Write(row)
	}
}

// main sets up the HTTP server and its routes
func main() {
	// Serve static files from the ./static directory
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Set up API endpoints
	http.HandleFunc("/scan", scanHandler)
	http.HandleFunc("/export/json", exportJSONHandler)
	http.HandleFunc("/export/csv", exportCSVHandler)

	// Start the server
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
