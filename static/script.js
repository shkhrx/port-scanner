// Get references to DOM elements
const form = document.getElementById('scanForm');
const resultsDiv = document.getElementById('results');
const loadingOverlay = document.getElementById('loadingOverlay');
const geoipDiv = document.getElementById('geoip');
const exportJsonBtn = document.getElementById('exportJson');
const exportCsvBtn = document.getElementById('exportCsv');

// Track if the last scan had any results
let lastScanHasResults = false;

// Handle form submission
form.addEventListener('submit', async function(e) {
    e.preventDefault();
    // Clear previous results
    resultsDiv.innerHTML = "";
    geoipDiv.innerHTML = "";
    loadingOverlay.classList.remove("hidden");
    lastScanHasResults = false;

    // Get form values
    const target = document.getElementById('target').value;
    const start = parseInt(document.getElementById('start').value);
    const end = parseInt(document.getElementById('end').value);

    try {
        // Send scan request to backend
        const res = await fetch('/scan', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({target, start, end})
        });
        if (!res.ok) throw new Error("Server error");
        const data = await res.json();

        // Display GeoIP information if available
        if (data.geoip) {
            geoipDiv.innerHTML = `
                <div class="mb-2 font-semibold">GeoIP Info:</div>
                <div>
                    <span class="font-semibold">IP:</span> ${data.geoip.query} &nbsp;|&nbsp;
                    <span class="font-semibold">Country:</span> ${data.geoip.country} &nbsp;|&nbsp;
                    <span class="font-semibold">Region:</span> ${data.geoip.regionName} &nbsp;|&nbsp;
                    <span class="font-semibold">City:</span> ${data.geoip.city} <br>
                    <span class="font-semibold">ISP:</span> ${data.geoip.isp} &nbsp;|&nbsp;
                    <span class="font-semibold">Org:</span> ${data.geoip.org}
                </div>
            `;
        }

        // Display scan results in a table
        if (data.ports && data.ports.length) {
            lastScanHasResults = true;
            let table = `
                <table class="mt-4">
                    <thead>
                        <tr>
                            <th>Port</th>
                            <th>Service</th>
                            <th>Response (ms)</th>
                            <th>Banner</th>
                        </tr>
                    </thead>
                    <tbody>
            `;
            // Add each port result to the table
            data.ports.forEach(port => {
                table += `
                    <tr>
                        <td class="result-blue font-bold text-lg">${port.port}</td>
                        <td class="font-semibold text-gray-700">${port.service || '-'}</td>
                        <td class="text-gray-600">${port.response_ms} ms</td>
                        <td class="banner-cell whitespace-pre-wrap break-all">${port.banner ? port.banner.replace(/</g, "&lt;").replace(/>/g, "&gt;").trim() : '-'}</td>
                    </tr>
                `;
            });
            table += `</tbody></table>`;
            resultsDiv.innerHTML = table;
        } else {
            // Display message if no open ports found
            resultsDiv.innerHTML = `<div class="error-red">No open ports found.</div>`;
        }
    } catch (err) {
        // Display error message if scan fails
        resultsDiv.innerHTML = `<div class="error-red">Error: ${err.message}</div>`;
    } finally {
        // Hide loading overlay
        loadingOverlay.classList.add("hidden");
    }
});

// Handle JSON export button click
exportJsonBtn.addEventListener('click', () => {
    if (!lastScanHasResults) return;
    window.location = '/export/json';
});

// Handle CSV export button click
exportCsvBtn.addEventListener('click', () => {
    if (!lastScanHasResults) return;
    window.location = '/export/csv';
});