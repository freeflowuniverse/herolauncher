/**
 * Utility functions for LiveKit Meet
 */

// Helper function to extract the original name from participant identity
// Removes the random suffix added by the server (format: name__randomString)
function getOriginalName(identity) {
  if (!identity) return '';
  
  // Split by the double underscore and take the first part
  const parts = identity.split('__');
  return parts[0];
}

// Encode passphrase for URL fragment
function encodePassphrase(passphrase) {
  return encodeURIComponent(passphrase);
}

// Decode passphrase from URL fragment
function decodePassphrase(encodedString) {
  if (!encodedString) return '';
  return decodeURIComponent(encodedString);
}

// Generate a random room ID
function generateRoomId() {
  return `${randomString(4)}-${randomString(4)}`;
}

// Generate a random string of specified length
function randomString(length) {
  let result = '';
  const characters = 'abcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return result;
}

// Format chat message links
function formatChatMessageLinks(message) {
  const urlRegex = /(https?:\/\/[^\s]+)/g;
  return message.replace(urlRegex, url => `<a href="${url}" target="_blank" rel="noopener noreferrer">${url}</a>`);
}

// Add a chat message to the chat panel
function addChatMessage(sender, content, isLocal = false) {
  const chatMessages = document.getElementById('chat-messages');
  const messageElement = document.createElement('div');
  messageElement.className = `chat-message ${isLocal ? 'local' : 'remote'}`;
  
  const senderElement = document.createElement('div');
  senderElement.className = 'sender';
  // Use original name without random suffix
  senderElement.textContent = isLocal ? 'You' : getOriginalName(sender);
  
  const contentElement = document.createElement('div');
  contentElement.className = 'content';
  contentElement.innerHTML = formatChatMessageLinks(content);
  
  messageElement.appendChild(senderElement);
  messageElement.appendChild(contentElement);
  chatMessages.appendChild(messageElement);
  
  // Scroll to bottom
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Check if a codec is a valid video codec
function isVideoCodec(codec) {
  const validCodecs = ['vp8', 'vp9', 'h264', 'av1'];
  return validCodecs.includes(codec);
}
