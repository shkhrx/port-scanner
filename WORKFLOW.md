# Port Scanner Workflow Documentation

## Overview
This document explains the chronological workflow of the port scanner application, from user input to result display.

## 1. Frontend Initialization
- The application loads with three main components:
  - Left sidebar: Common IPs and ports reference
  - Main content: Scan form and results
  - Right sidebar: Information about port scanning

## 2. User Input Process
1. User enters:
   - Target (IP address or domain)
   - Start port number
   - End port number
2. User clicks "Scan" button
3. Frontend JavaScript:
   - Prevents default form submission
   - Clears previous results
   - Shows loading overlay
   - Prepares data for API request

## 3. API Request Flow
1. Frontend sends POST request to `/scan` endpoint:
   ```javascript
   fetch('/scan', {
       method: 'POST',
       headers: {'Content-Type': 'application/json'},
       body: JSON.stringify({target, start, end})
   })
   ```

2. Backend (Go) receives request:
   - Decodes JSON request body into `ScanRequest` struct
   - Validates input parameters
   - Initiates GeoIP lookup

## 4. GeoIP Lookup Process
1. Backend makes HTTP request to ip-api.com:
   ```go
   resp, err := http.Get("http://ip-api.com/json/" + target)
   ```
2. Decodes response into `GeoIPInfo` struct
3. Stores information for later use in results

## 5. Port Scanning Process
1. Backend creates synchronization primitives:
   - WaitGroup for concurrent operations
   - Mutex for thread-safe result collection
   - Semaphore to limit concurrent connections (100 max)

2. For each port in range:
   - Launches goroutine for concurrent scanning
   - Attempts TCP connection with 1-second timeout
   - If connection successful:
     * Records response time
     * Attempts to read banner (500ms timeout)
     * Identifies service from common services map
     * Stores port details

3. Waits for completion or timeout (30 seconds)

## 6. Result Processing
1. Backend:
   - Collects all successful port scans
   - Combines with GeoIP information
   - Creates `ScanResult` struct
   - Stores result for potential export
   - Sends JSON response to frontend

2. Frontend receives response:
   - Parses JSON data
   - Updates UI with results

## 7. Results Display
1. GeoIP Information:
   - Displays target location details
   - Shows ISP and organization information

2. Port Results Table:
   - Creates table with headers
   - Adds row for each open port
   - Displays:
     * Port number
     * Service name
     * Response time
     * Banner information

## 8. Export Functionality
1. JSON Export:
   - User clicks "Export as JSON"
   - Frontend redirects to `/export/json`
   - Backend retrieves last scan result
   - Sends JSON file download

2. CSV Export:
   - User clicks "Export as CSV"
   - Frontend redirects to `/export/csv`
   - Backend:
     * Creates CSV with headers
     * Adds rows with GeoIP and port data
     * Sends CSV file download

## 9. Error Handling
1. Frontend:
   - Network errors
   - Invalid input validation
   - Display error messages
   - Loading state management

2. Backend:
   - Connection timeouts
   - Invalid requests
   - GeoIP lookup failures
   - Concurrent operation safety

## 10. Performance Considerations
1. Concurrency:
   - Limited to 100 concurrent connections
   - Uses goroutines for parallel scanning
   - Mutex for thread-safe operations

2. Timeouts:
   - Connection: 1 second
   - Banner read: 500ms
   - Overall scan: 30 seconds

3. Resource Management:
   - Proper connection closing
   - Memory-efficient data structures
   - Controlled concurrent operations

## 11. Security Measures
1. Input Validation:
   - Port range validation
   - Target format checking
   - Request size limits

2. Connection Safety:
   - Timeout limits
   - Connection limits
   - Proper resource cleanup

3. Export Security:
   - File type validation
   - Size limits
   - Proper headers

## 12. User Experience Flow
1. Input Phase:
   - User enters scan parameters
   - Form validation
   - Loading indication

2. Processing Phase:
   - Loading overlay
   - Progress indication
   - Error handling

3. Results Phase:
   - Clear data presentation
   - Export options
   - Error messages if needed

This workflow ensures efficient, safe, and user-friendly port scanning with proper error handling and resource management. 