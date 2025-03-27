// Variables for logs functionality
let currentServiceName = '';

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
  currentServiceName = '';
}

// Function to fetch process logs
function fetchProcessLogs(name, lines = 50) {
  const formData = new FormData();
  formData.append('name', name);
  formData.append('lines', lines);
  
  const logsContent = document.getElementById('logs-content');
  if (!logsContent) return;
  
  logsContent.textContent = 'Loading logs...';
  
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
        logsContent.textContent = cleanedLogs;
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
