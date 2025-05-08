const form = document.getElementById('scanForm');
const resultsDiv = document.getElementById('results');

form.addEventListener('submit', async function(e) {
    e.preventDefault();
    resultsDiv.textContent = "Scanning...";
    resultsDiv.classList.remove("text-green-600", "text-red-600");
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
        if (data.open_ports.length) {
            resultsDiv.textContent = 'Open ports: ' + data.open_ports.join(', ');
            resultsDiv.classList.add("text-green-600");
        } else {
            resultsDiv.textContent = 'No open ports found.';
            resultsDiv.classList.add("text-red-600");
        }
    } catch (err) {
        resultsDiv.textContent = "Error: " + err.message;
        resultsDiv.classList.add("text-red-600");
    }
});