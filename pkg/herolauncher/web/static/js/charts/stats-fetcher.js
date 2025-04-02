// Data fetching functions for system stats
document.addEventListener('DOMContentLoaded', function() {
  // Function to fetch hardware stats
  function fetchHardwareStats() {
    fetch('/api/hardware-stats')
      .then(response => {
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        return response.json();
      })
      .then(data => {
        // Extract network speeds
        var upSpeed = data.network && data.network.upload_speed ? data.network.upload_speed : '0Mbps';
        var downSpeed = data.network && data.network.download_speed ? data.network.download_speed : '0Mbps';
        
        // Update the network chart
        if (window.updateNetworkChart) {
          window.updateNetworkChart(upSpeed, downSpeed);
        }
      })
      .catch(error => {
        console.error('Error fetching hardware stats:', error);
      });
  }
  
  // Function to fetch process stats
  function fetchProcessStats() {
    fetch('/api/process-stats')
      .then(response => {
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        return response.json();
      })
      .then(data => {
        // Update the CPU and Memory charts with new data
        if (window.updateCpuChart && data.processes) {
          window.updateCpuChart(data.processes);
        }
        if (window.updateMemoryChart && data.processes) {
          window.updateMemoryChart(data.processes);
        }
      })
      .catch(error => {
        console.error('Error fetching process stats:', error);
      });
  }
  
  // Function to fetch all stats
  function fetchAllStats() {
    fetchHardwareStats();
    fetchProcessStats();
    
    // Schedule the next update - use requestAnimationFrame for smoother updates
    requestAnimationFrame(function() {
      setTimeout(fetchAllStats, 2000); // Update every 2 seconds
    });
  }
  
  // Start fetching all stats if we're on the system info page
  if (document.getElementById('cpu-chart') || 
      document.getElementById('memory-chart') || 
      document.getElementById('network-chart')) {
    fetchAllStats();
  }
  
  // Also update the chart when new hardware stats are loaded via Unpoly
  document.addEventListener('up:fragment:loaded', function(event) {
    if (event.target && event.target.classList.contains('hardware-stats')) {
      // Extract network speeds from the table
      var networkCell = event.target.querySelector('tr:nth-child(4) td');
      if (networkCell) {
        var networkText = networkCell.textContent;
        var upMatch = networkText.match(/Up: ([\d.]+Mbps)/);
        var downMatch = networkText.match(/Down: ([\d.]+Mbps)/);
        
        var upSpeed = upMatch ? upMatch[1] : '0Mbps';
        var downSpeed = downMatch ? downMatch[1] : '0Mbps';
        
        // Update the chart with new data
        if (window.updateNetworkChart) {
          window.updateNetworkChart(upSpeed, downSpeed);
        }
      }
    }
  });
});
