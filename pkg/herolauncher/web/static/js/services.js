// Function to refresh services
function refreshServices() {
    const servicesTable = document.getElementById('services-table');
    fetch('/admin/services/data')
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.error || 'Failed to refresh services');
                });
            }
            return response.text();
        })
        .then(html => {
            servicesTable.innerHTML = html;
        })
        .catch(error => {
            console.error('Error refreshing services:', error);
            // Show error message in the services table instead of replacing it
            const errorHtml = `<table><tbody><tr><td colspan="4"><div class="alert alert-danger">Error refreshing services: ${error.message}</div></td></tr></tbody></table>`;
            servicesTable.innerHTML = errorHtml;
            // Try again after a short delay
            setTimeout(() => {
                refreshServices();
            }, 3000);
        });
}

// Refresh services as soon as the page loads
document.addEventListener('DOMContentLoaded', function() {
    refreshServices();
});

// Function to start a new service
function startService(event) {
    event.preventDefault();
    const form = document.getElementById('start-service-form');
    const resultDiv = document.getElementById('start-result');
    
    const formData = new FormData(form);
    
    fetch('/admin/services/start', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                resultDiv.className = 'alert alert-danger';
                resultDiv.textContent = data.error;
            } else {
                resultDiv.className = 'alert alert-success';
                resultDiv.textContent = data.message;
                form.reset();
                refreshServices();
            }
            resultDiv.style.display = 'block';
            setTimeout(() => {
                resultDiv.style.display = 'none';
            }, 5000);
        })
        .catch(error => {
            resultDiv.className = 'alert alert-danger';
            resultDiv.textContent = 'An error occurred: ' + error.message;
            resultDiv.style.display = 'block';
        });
}

// Function to stop a process
function stopProcess(name) {
    if (!confirm('Are you sure you want to stop this service?')) return;
    
    const formData = new FormData();
    formData.append('name', name);
    
    fetch('/admin/services/stop', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                alert('Error: ' + data.error);
            } else {
                refreshServices();
            }
        })
        .catch(error => {
            alert('An error occurred: ' + error.message);
        });
}

// Function to restart a process
function restartProcess(name) {
    if (!confirm('Are you sure you want to restart this service?')) return;
    
    const formData = new FormData();
    formData.append('name', name);
    
    fetch('/admin/services/restart', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                alert('Error: ' + data.error);
            } else {
                refreshServices();
            }
        })
        .catch(error => {
            alert('An error occurred: ' + error.message);
        });
}

// Function to delete a process
function deleteProcess(name) {
    if (!confirm('Are you sure you want to delete this service? This cannot be undone.')) return;
    
    const formData = new FormData();
    formData.append('name', name);
    
    fetch('/admin/services/delete', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                alert('Error: ' + data.error);
            } else {
                refreshServices();
            }
        })
        .catch(error => {
            alert('An error occurred: ' + error.message);
        });
}

// Function to show process logs
function showProcessLogs(name) {
    // Create a modal to show logs
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>Logs for ${name}</h2>
                <span class="close">&times;</span>
            </div>
            <div class="modal-body">
                <pre id="log-content" style="height: 400px; overflow-y: auto; background: #f5f5f5; padding: 10px;">Loading logs...</pre>
            </div>
            <div class="modal-footer">
                <button class="button refresh" onclick="refreshLogs('${name}')">Refresh Logs</button>
                <button class="button secondary" onclick="closeModal()">Close</button>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
    
    // Add modal styles if not already present
    if (!document.getElementById('modal-styles')) {
        const style = document.createElement('style');
        style.id = 'modal-styles';
        style.innerHTML = `
            .modal {
                display: block;
                position: fixed;
                z-index: 1000;
                left: 0;
                top: 0;
                width: 100%;
                height: 100%;
                background-color: rgba(0,0,0,0.4);
            }
            .modal-content {
                background-color: #fefefe;
                margin: 5% auto;
                padding: 20px;
                border: 1px solid #888;
                width: 80%;
                max-width: 800px;
                border-radius: 5px;
                box-shadow: 0 4px 8px rgba(0,0,0,0.1);
            }
            .modal-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                border-bottom: 1px solid #eee;
                padding-bottom: 10px;
                margin-bottom: 15px;
            }
            .modal-footer {
                border-top: 1px solid #eee;
                padding-top: 15px;
                margin-top: 15px;
                text-align: right;
            }
            .close {
                color: #aaa;
                font-size: 28px;
                font-weight: bold;
                cursor: pointer;
            }
            .close:hover {
                color: black;
            }
        `;
        document.head.appendChild(style);
    }
    
    // Close modal when clicking the X
    modal.querySelector('.close').onclick = closeModal;
    
    // Load the logs
    loadLogs(name);
    
    // Close modal when clicking outside
    window.onclick = function(event) {
        if (event.target === modal) {
            closeModal();
        }
    };
}

// Function to load logs
function loadLogs(name) {
    fetch(`/admin/services/logs?name=${encodeURIComponent(name)}&lines=100`)
        .then(response => response.json())
        .then(data => {
            const logContent = document.getElementById('log-content');
            if (data.error) {
                logContent.textContent = `Error: ${data.error}`;
            } else {
                logContent.textContent = data.logs || 'No logs available';
                // Scroll to bottom
                logContent.scrollTop = logContent.scrollHeight;
            }
        })
        .catch(error => {
            document.getElementById('log-content').textContent = `Error loading logs: ${error.message}`;
        });
}

// Function to refresh logs
function refreshLogs(name) {
    document.getElementById('log-content').textContent = 'Refreshing logs...';
    loadLogs(name);
}

// Function to close the modal
function closeModal() {
    const modal = document.querySelector('.modal');
    if (modal) {
        document.body.removeChild(modal);
    }
    window.onclick = null;
}
