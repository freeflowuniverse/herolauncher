// Initialize Unpoly
up.log.enable();

// Configure Unpoly
up.fragment.config.mainTargets = ['.app-content'];

// Add a transition effect when loading content
up.transition.config.duration = 300;

// Document ready event
document.addEventListener('DOMContentLoaded', function() {
  initializeApp();
});

// Initialize the application
function initializeApp() {
  // Initialize room functionality if we're in a conference room
  const conferenceRoom = document.querySelector('.conference-room');
  if (conferenceRoom) {
    initializeRoom(conferenceRoom);
  }
  
  // Initialize create room functionality
  const createButton = document.querySelector('.create-button');
  if (createButton) {
    createButton.addEventListener('click', function() {
      // Generate a random room ID
      const roomId = generateRoomId();
      // Redirect to the join form with the room ID pre-filled
      document.getElementById('roomId').value = roomId;
    });
  }
}

// Initialize room functionality
function initializeRoom(roomElement) {
  const roomId = roomElement.dataset.roomId;
  
  // Initialize tabs
  initializeTabs();
  
  // Initialize control buttons
  initializeControls();
  
  // Initialize leave button
  const leaveButton = document.querySelector('.leave-button');
  if (leaveButton) {
    leaveButton.addEventListener('click', function() {
      if (confirm('Are you sure you want to leave this room?')) {
        window.location.href = '/';
      }
    });
  }
  
  // In a real application, this is where you would initialize the video conferencing SDK
  // For example, connecting to a WebRTC service or a video conferencing platform
  console.log(`Initializing room ${roomId}`);
  
  // Simulate participants joining (for demo purposes)
  simulateParticipants();
}

// Initialize tabs in the room sidebar
function initializeTabs() {
  const tabButtons = document.querySelectorAll('.tab-button');
  const tabPanes = document.querySelectorAll('.tab-pane');
  
  tabButtons.forEach(button => {
    button.addEventListener('click', function() {
      // Remove active class from all buttons and panes
      tabButtons.forEach(btn => btn.classList.remove('active'));
      tabPanes.forEach(pane => pane.classList.remove('active'));
      
      // Add active class to clicked button
      this.classList.add('active');
      
      // Show corresponding tab pane
      const tabId = this.dataset.tab;
      document.getElementById(`${tabId}-tab`).classList.add('active');
    });
  });
}

// Initialize control buttons
function initializeControls() {
  const muteAudioButton = document.getElementById('mute-audio');
  const muteVideoButton = document.getElementById('mute-video');
  const shareScreenButton = document.getElementById('share-screen');
  const settingsButton = document.getElementById('settings');
  
  if (muteAudioButton) {
    muteAudioButton.addEventListener('click', function() {
      this.classList.toggle('muted');
      // In a real application, this would toggle the audio track
      console.log('Toggle audio mute');
    });
  }
  
  if (muteVideoButton) {
    muteVideoButton.addEventListener('click', function() {
      this.classList.toggle('muted');
      // In a real application, this would toggle the video track
      console.log('Toggle video mute');
    });
  }
  
  if (shareScreenButton) {
    shareScreenButton.addEventListener('click', function() {
      this.classList.toggle('sharing');
      // In a real application, this would start/stop screen sharing
      console.log('Toggle screen sharing');
    });
  }
  
  if (settingsButton) {
    settingsButton.addEventListener('click', function() {
      // In a real application, this would open a settings dialog
      console.log('Open settings');
    });
  }
}

// Generate a random room ID
function generateRoomId() {
  return Math.random().toString(36).substring(2, 8).toUpperCase();
}

// Simulate participants joining (for demo purposes)
function simulateParticipants() {
  const participantsList = document.getElementById('participants-list');
  const participantsGrid = document.getElementById('participants-grid');
  
  if (!participantsList || !participantsGrid) return;
  
  // Demo participants
  const demoParticipants = [
    { id: 'user1', name: 'Alice Smith' },
    { id: 'user2', name: 'Bob Johnson' },
    { id: 'user3', name: 'Carol Davis' }
  ];
  
  // Add participants to the list and grid
  demoParticipants.forEach(participant => {
    // Add to participants list
    const listItem = document.createElement('li');
    listItem.innerHTML = `
      <span class="participant-name">${participant.name}</span>
    `;
    participantsList.appendChild(listItem);
    
    // Add to participants grid
    const videoElement = document.createElement('div');
    videoElement.className = 'video-container';
    videoElement.innerHTML = `
      <div class="placeholder-video">
        <p>${participant.name}</p>
      </div>
    `;
    participantsGrid.appendChild(videoElement);
  });
  
  // Simulate chat messages
  simulateChat();
}

// Simulate chat messages (for demo purposes)
function simulateChat() {
  const chatMessages = document.getElementById('chat-messages');
  const chatForm = document.getElementById('chat-form');
  
  if (!chatMessages || !chatForm) return;
  
  // Demo messages
  const demoMessages = [
    { sender: 'Alice Smith', message: 'Hello everyone!' },
    { sender: 'Bob Johnson', message: 'Hi Alice, how are you?' },
    { sender: 'Carol Davis', message: 'Good morning team!' }
  ];
  
  // Add demo messages to chat
  demoMessages.forEach(msg => {
    const messageElement = document.createElement('div');
    messageElement.className = 'chat-message';
    messageElement.innerHTML = `
      <strong>${msg.sender}:</strong> ${msg.message}
    `;
    chatMessages.appendChild(messageElement);
  });
  
  // Handle chat form submission
  chatForm.addEventListener('submit', function(e) {
    e.preventDefault();
    
    const input = document.getElementById('chat-input');
    const message = input.value.trim();
    
    if (message) {
      // Add message to chat
      const messageElement = document.createElement('div');
      messageElement.className = 'chat-message';
      messageElement.innerHTML = `
        <strong>You:</strong> ${message}
      `;
      chatMessages.appendChild(messageElement);
      
      // Clear input
      input.value = '';
      
      // Scroll to bottom
      chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  });
}
