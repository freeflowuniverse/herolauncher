// Variables for logs functionality
let currentServiceName = '';
let autoRefreshEnabled = false;
let autoRefreshInterval = null;
const AUTO_REFRESH_RATE = 3000; // 3 seconds

// Function to show process logs
function showProcessLogs(name) {
  currentServiceName = name;
  
  // Create modal if it doesn't exist
  let modal = document.getElementById('logs-modal');
  if (!modal) {
    modal = createLogsModal();
  }
  
  document.getElementById('logs-modal-title').textContent = `Service Logs: ${name}`;
  modal.style.display = 'block';
  fetchProcessLogs(name);
}

// Function to create the logs modal
function createLogsModal() {
  const modal = document.createElement('div');
  modal.id = 'logs-modal';
  modal.className = 'modal';
  modal.style.display = 'none';
  modal.innerHTML = `
    <div class="modal-background" onclick="closeLogsModal()"></div>
    <div class="modal-content">
      <div class="modal-header">
        <h3 id="logs-modal-title">Service Logs</h3>
        <span class="close" onclick="closeLogsModal()">&times;</span>
      </div>
      <div class="modal-body">
        <pre id="logs-content">Loading logs...</pre>
      </div>
      <div class="modal-footer">
        <label class="auto-refresh-toggle">
          <input type="checkbox" id="auto-refresh-checkbox" onchange="toggleAutoRefresh()">
          <span>Auto-refresh</span>
        </label>
        <button class="button secondary" onclick="closeLogsModal()">Close</button>
        <button class="button primary" onclick="refreshLogs()">Refresh</button>
      </div>
    </div>
  `;
  document.body.appendChild(modal);
  
  // Add modal styles
  const style = document.createElement('style');
  style.textContent = `
    .modal {
      display: none;
      position: fixed;
      z-index: 1000;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      overflow: auto;
      background-color: rgba(0,0,0,0.4);
    }
    
    .modal-content {
      background-color: #fefefe;
      margin: 10% auto;
      padding: 0;
      border: 1px solid #888;
      width: 80%;
      max-width: 800px;
      box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
      border-radius: 4px;
    }
    
    .modal-header {
      padding: 10px 15px;
      background-color: #f8f9fa;
      border-bottom: 1px solid #dee2e6;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    
    .modal-header h3 {
      margin: 0;
    }
    
    .close {
      color: #aaa;
      font-size: 28px;
      font-weight: bold;
      cursor: pointer;
    }
    
    .close:hover,
    .close:focus {
      color: black;
      text-decoration: none;
    }
    
    .modal-body {
      padding: 15px;
      max-height: 500px;
      overflow-y: auto;
    }
    
    .modal-body pre {
      white-space: pre-wrap;
      word-wrap: break-word;
      background-color: #f8f9fa;
      padding: 10px;
      border-radius: 4px;
      border: 1px solid #dee2e6;
      font-family: monospace;
      margin: 0;
      height: 400px;
      overflow-y: auto;
    }
    
    .modal-footer {
      padding: 10px 15px;
      background-color: #f8f9fa;
      border-top: 1px solid #dee2e6;
      display: flex;
      justify-content: flex-end;
      gap: 10px;
    }
    
    .auto-refresh-toggle {
      display: flex;
      align-items: center;
      margin-right: auto;
      cursor: pointer;
    }
    
    .auto-refresh-toggle input {
      margin-right: 5px;
    }
  `;
  document.head.appendChild(style);
  
  return modal;
}

// Function to close the logs modal
function closeLogsModal() {
  const modal = document.getElementById('logs-modal');
  if (modal) {
    modal.style.display = 'none';
  }
  
  // Disable auto-refresh when closing the modal
  disableAutoRefresh();
  currentServiceName = '';
}

// Function to fetch process logs
function fetchProcessLogs(name, lines = 10000) {
  const formData = new FormData();
  formData.append('name', name);
  formData.append('lines', lines);
  
  const logsContent = document.getElementById('logs-content');
  if (!logsContent) return;
  
  // Save scroll position if auto-refreshing
  const isAutoRefresh = autoRefreshEnabled;
  const scrollTop = isAutoRefresh ? logsContent.scrollTop : 0;
  const scrollHeight = isAutoRefresh ? logsContent.scrollHeight : 0;
  const clientHeight = isAutoRefresh ? logsContent.clientHeight : 0;
  const wasScrolledToBottom = scrollHeight - scrollTop <= clientHeight + 5; // 5px tolerance
  
  // Only show loading indicator on first load, not during auto-refresh
  if (!isAutoRefresh) {
    logsContent.textContent = 'Loading logs...';
  }
  
  fetch('/admin/services/logs', {
    method: 'POST',
    body: formData
  })
    .then(response => response.json())
    .then(data => {
      if (data.error) {
        logsContent.textContent = `Error: ${data.error}`;
      } else {
        // Clean up the logs by removing **RESULT** and **ENDRESULT** markers
        let cleanedLogs = data.logs || 'No logs available';
        cleanedLogs = cleanedLogs.replace(/\*\*RESULT\*\*/g, '');
        cleanedLogs = cleanedLogs.replace(/\*\*ENDRESULT\*\*/g, '');
        // Trim extra whitespace
        cleanedLogs = cleanedLogs.trim();
        
        // Format the logs with stderr lines in red
        if (cleanedLogs.length > 0) {
          // Clear the logs content
          logsContent.textContent = '';
          
          // Split the logs into lines and process each line
          const lines = cleanedLogs.split('\n');
          lines.forEach(line => {
            const logLine = document.createElement('div');
            
            // Check if this is a stderr line (starts with timestamp followed by E)
            if (line.match(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} E /)) {
              logLine.className = 'stderr-log';
              logLine.style.color = '#ff3333'; // Red color for stderr
            }
            
            logLine.textContent = line;
            logsContent.appendChild(logLine);
          });
          
          // Add some styling for the pre element to maintain formatting
          logsContent.style.fontFamily = 'monospace';
          logsContent.style.whiteSpace = 'pre-wrap';
          
          // Restore scroll position if it was an auto-refresh
          if (isAutoRefresh) {
            if (wasScrolledToBottom) {
              // If user was at the bottom, keep them at the bottom
              logsContent.scrollTop = logsContent.scrollHeight;
            } else {
              // Otherwise maintain the same scroll position
              logsContent.scrollTop = scrollTop;
            }
          }
        } else {
          logsContent.textContent = 'No logs available';
        }
      }
    })
    .catch(error => {
      logsContent.textContent = `Error loading logs: ${error.message}`;
    });
}

// Function to refresh logs for the current service
function refreshLogs() {
  if (currentServiceName) {
    fetchProcessLogs(currentServiceName);
  }
}

// Function to toggle auto-refresh
function toggleAutoRefresh() {
  const checkbox = document.getElementById('auto-refresh-checkbox');
  if (checkbox && checkbox.checked) {
    enableAutoRefresh();
  } else {
    disableAutoRefresh();
  }
}

// Function to enable auto-refresh
function enableAutoRefresh() {
  // Don't create multiple intervals
  if (autoRefreshInterval) {
    clearInterval(autoRefreshInterval);
  }
  
  // Set the flag
  autoRefreshEnabled = true;
  
  // Create the interval
  autoRefreshInterval = setInterval(() => {
    if (currentServiceName) {
      fetchProcessLogs(currentServiceName);
    }
  }, AUTO_REFRESH_RATE);
  
  console.log('Auto-refresh enabled with interval:', AUTO_REFRESH_RATE, 'ms');
}

// Function to disable auto-refresh
function disableAutoRefresh() {
  autoRefreshEnabled = false;
  
  if (autoRefreshInterval) {
    clearInterval(autoRefreshInterval);
    autoRefreshInterval = null;
  }
  
  // Uncheck the checkbox if it exists
  const checkbox = document.getElementById('auto-refresh-checkbox');
  if (checkbox) {
    checkbox.checked = false;
  }
  
  console.log('Auto-refresh disabled');
}

// Close modal when clicking outside of it
window.addEventListener('click', function(event) {
  const modal = document.getElementById('logs-modal');
  if (modal && event.target === modal) {
    closeLogsModal();
  }
});

// Allow ESC key to close the modal
document.addEventListener('keydown', function(event) {
  if (event.key === 'Escape') {
    closeLogsModal();
  }
});
