const form = document.getElementById('scanForm');
const resultsDiv = document.getElementById('results');
const loadingOverlay = document.getElementById('loadingOverlay');
const geoipDiv = document.getElementById('geoip');
const exportJsonBtn = document.getElementById('exportJson');
const exportCsvBtn = document.getElementById('exportCsv');

let lastScanHasResults = false;

form.addEventListener('submit', async function(e) {
    e.preventDefault();
    resultsDiv.innerHTML = "";
    geoipDiv.innerHTML = "";
    loadingOverlay.classList.remove("hidden");
    lastScanHasResults = false;

    const target = document.getElementById('target').value;
    const start = parseInt(document.getElementById('start').value);
    const end = parseInt(document.getElementById('end').value);

    try {
        const res = await fetch('/scan', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({target, start, end})
        });
        if (!res.ok) throw new Error("Server error");
        const data = await res.json();

        // GeoIP Info
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

        // Results Table
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
            data.ports.forEach(port => {
                table += `
                    <tr>
                        <td class="result-blue font-bold text-lg">${port.port}</td>
                        <td>${port.service || '-'}</td>
                        <td>${port.response_ms}</td>
                        <td class="banner-cell">${port.banner ? port.banner.replace(/</g, "&lt;").replace(/>/g, "&gt;") : '-'}</td>
                    </tr>
                `;
            });
            table += `</tbody></table>`;
            resultsDiv.innerHTML = table;
        } else {
            resultsDiv.innerHTML = `<div class="error-red">No open ports found.</div>`;
        }
    } catch (err) {
        resultsDiv.innerHTML = `<div class="error-red">Error: ${err.message}</div>`;
    } finally {
        loadingOverlay.classList.add("hidden");
    }
});

// Export buttons
exportJsonBtn.addEventListener('click', () => {
    if (!lastScanHasResults) return;
    window.location = '/export/json';
});
exportCsvBtn.addEventListener('click', () => {
    if (!lastScanHasResults) return;
    window.location = '/export/csv';
});