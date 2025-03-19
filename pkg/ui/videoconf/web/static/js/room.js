document.addEventListener('DOMContentLoaded', async () => {
    // Get DOM elements
    const preJoinContainer = document.getElementById('pre-join-container');
    const conferenceContainer = document.getElementById('conference-container');
    const preJoinForm = document.getElementById('pre-join-form');
    const usernameInput = document.getElementById('username');
    const videoEnabledCheckbox = document.getElementById('video-enabled');
    const audioEnabledCheckbox = document.getElementById('audio-enabled');
    const localVideo = document.getElementById('local-video');
    const remoteParticipants = document.getElementById('remote-participants');
    const toggleVideoBtn = document.getElementById('toggle-video');
    const toggleAudioBtn = document.getElementById('toggle-audio');
    const shareScreenBtn = document.getElementById('share-screen');
    const toggleChatBtn = document.getElementById('toggle-chat');
    const leaveRoomBtn = document.getElementById('leave-room');
    const chatPanel = document.getElementById('chat-panel');
    const closeChatBtn = document.getElementById('close-chat');
    const chatForm = document.getElementById('chat-form');
    const chatMessageInput = document.getElementById('chat-message');
    
    // Get room information from window variables
    const roomName = window.ROOM_NAME;
    const region = window.REGION;
    const hq = window.HQ;
    const codec = window.CODEC;
    
    // Check for E2EE passphrase in URL fragment
    const e2eePassphrase = decodePassphrase(window.location.hash.substring(1));
    const e2eeEnabled = !!e2eePassphrase;
    
    let room;
    let localParticipant;
    
    // Handle pre-join form submission
    preJoinForm.addEventListener('submit', async (event) => {
      event.preventDefault();
      
      const username = usernameInput.value.trim();
      if (!username) return;
      
      const videoEnabled = videoEnabledCheckbox.checked;
      const audioEnabled = audioEnabledCheckbox.checked;
      
      try {
        // Get connection details from server
        const url = new URL('/api/connection-details', window.location.origin);
        url.searchParams.append('roomName', roomName);
        url.searchParams.append('participantName', username);
        if (region) {
          url.searchParams.append('region', region);
        }
        
        const response = await fetch(url.toString());
        if (!response.ok) {
          throw new Error('Failed to get connection details');
        }
        
        const connectionDetails = await response.json();
        
        // Debug the connection details
        console.log('Connection details from server:', connectionDetails);
        
        // Validate that we have a proper token
        if (!connectionDetails.participantToken || typeof connectionDetails.participantToken !== 'string' || connectionDetails.participantToken.trim() === '') {
          throw new Error('Invalid token received from server');
        }
        
        // Connect to LiveKit room
        await connectToRoom(connectionDetails, {
          videoEnabled,
          audioEnabled
        });
        
        // Show conference UI
        preJoinContainer.style.display = 'none';
        conferenceContainer.style.display = 'block';
      } catch (error) {
        console.error('Error joining room:', error);
        alert(`Error joining room: ${error.message}`);
      }
    });
    
    // Connect to LiveKit room
    async function connectToRoom(connectionDetails, userChoices) {
      try {
        console.log('Debug: Starting connectToRoom with details:', JSON.stringify(connectionDetails));
        console.log('Debug: User choices:', JSON.stringify(userChoices));
        // Create room options
        const roomOptions = {
          adaptiveStream: true,
          dynacast: true,
          videoCaptureDefaults: {
            resolution: hq ? LivekitClient.VideoPresets.h2160 : LivekitClient.VideoPresets.h720
          },
          publishDefaults: {
            simulcast: true,
            videoSimulcastLayers: hq 
              ? [LivekitClient.VideoPresets.h1080, LivekitClient.VideoPresets.h720]
              : [LivekitClient.VideoPresets.h540, LivekitClient.VideoPresets.h216],
            videoCodec: codec
          }
        };
        
        // Setup E2EE if enabled
        if (e2eeEnabled) {
          const worker = new Worker('/js/e2ee-worker.js');
          const keyProvider = new LivekitClient.ExternalE2EEKeyProvider();
          await keyProvider.setKey(e2eePassphrase);
          
          roomOptions.e2ee = {
            keyProvider,
            worker
          };
        }
        
        // Create room
        room = new LivekitClient.Room(roomOptions);
        
        // Set up event listeners
        setupRoomEventListeners(room);
        
        // Add debugging to see what's being sent
        console.log('Connecting to LiveKit server with:', {
          serverUrl: connectionDetails.serverUrl,
          token: connectionDetails.participantToken
        });
        
        // Get the token directly from the connection details
        const token = connectionDetails.participantToken;
        
        // Ensure the server URL is properly formatted
        let serverUrl = connectionDetails.serverUrl;
        if (serverUrl && !serverUrl.startsWith('wss://')) {
          serverUrl = serverUrl.replace(/^https?:\/\//, 'wss://');
        }
        
        // Remove any double slashes except after the protocol
        serverUrl = serverUrl.replace(/(wss:\/\/)\/*(.+)/, '$1$2');
        
        console.log('Final connection parameters:', { 
          serverUrl, 
          token,
          tokenType: typeof token
        });
        
        // Connect to room with the token as a direct string parameter
        console.log('Debug: About to connect to room with URL:', serverUrl);
        console.log('Debug: Token length:', token ? token.length : 0);
        try {
          await room.connect(serverUrl, token);
          console.log('Debug: Connected to room successfully');
        } catch (connectError) {
          console.error('Debug: Error connecting to room:', connectError);
          throw connectError;
        }
        
        // Enable/disable local tracks based on user choices
        console.log('Debug: Getting local participant');
        localParticipant = room.localParticipant;
        console.log('Debug: Local participant:', localParticipant ? localParticipant.identity : 'undefined');
        
        console.log('Debug: Setting media state based on user choices:', 
          userChoices ? `video: ${userChoices.videoEnabled}, audio: ${userChoices.audioEnabled}` : 'undefined');
        try {
          // Detect Firefox browser
          const isFirefox = navigator.userAgent.toLowerCase().indexOf('firefox') > -1;
          console.log('Debug: Browser detected:', isFirefox ? 'Firefox' : 'Other');
          
          if (userChoices.videoEnabled) {
            console.log('Debug: Enabling camera and microphone');
            
            if (isFirefox) {
              // Firefox-specific approach
              console.log('Debug: Using Firefox-specific media approach');
              try {
                // First try to enable camera with specific constraints for Firefox
                await localParticipant.setCameraEnabled(true, {
                  resolution: LivekitClient.VideoPresets.h720,
                  facingMode: 'user'
                });
                console.log('Debug: Camera enabled successfully on Firefox');
              } catch (cameraError) {
                console.error('Debug: Firefox camera error:', cameraError);
              }
              
              try {
                // Then enable microphone separately
                await localParticipant.setMicrophoneEnabled(true);
                console.log('Debug: Microphone enabled successfully on Firefox');
              } catch (micError) {
                console.error('Debug: Firefox microphone error:', micError);
              }
            } else {
              // Standard approach for other browsers
              await localParticipant.enableCameraAndMicrophone();
              console.log('Debug: Camera and microphone enabled successfully');
            }
          } else {
            console.log('Debug: Disabling camera and microphone');
            // This should work the same on all browsers
            await localParticipant.disableCameraAndMicrophone();
            console.log('Debug: Camera and microphone disabled successfully');
          }
        } catch (mediaError) {
          console.error('Debug: Error setting media state:', mediaError);
          console.error('Debug: Media error details:', mediaError.message);
          console.error('Debug: Media error stack:', mediaError.stack);
        }
        
        // Render local participant
        console.log('Debug: About to render local participant');
        try {
          renderLocalParticipant();
          console.log('Debug: Local participant rendered successfully');
        } catch (renderError) {
          console.error('Debug: Error rendering local participant:', renderError);
        }
        
        // Handle existing participants in the room
        console.log('Debug: Checking for existing participants in the room');
        try {
          // In LiveKit, room.participants is a Map of participant sid -> participant
          console.log('Debug: Room object:', room);
          console.log('Debug: Room state:', room.state);
          console.log('Debug: Room participants:', room.participants);
          console.log('Debug: Room participants object type:', typeof room.participants);
          
          // Check if room has a getParticipants method (some versions of LiveKit use this)
          if (typeof room.getParticipants === 'function') {
            console.log('Debug: Room has getParticipants method, trying to use it');
            const participants = room.getParticipants();
            console.log('Debug: getParticipants result:', participants);
            console.log('Debug: getParticipants count:', participants ? participants.length : 0);
          }
          
          // Try to access participants in different ways
          console.log('Debug: Trying alternative ways to access participants');
          
          // Check if we can access remote participants through a different property
          console.log('Debug: Room properties:');
          for (const prop in room) {
            if (typeof room[prop] !== 'function') {
              console.log(`Debug: Room.${prop} type:`, typeof room[prop]);
            }
          }
          
          // Check if participants is a Map
          if (room.participants instanceof Map) {
            console.log('Debug: Participants is a Map with size:', room.participants.size);
            // Use Map methods to iterate
            room.participants.forEach((participant, sid) => {
              console.log(`Debug: Map participant ${sid}:`, participant.identity);
            });
          } else {
            console.log('Debug: Participants is NOT a Map');
          }
          
          // Check if any remote participants exist
          // Different versions of LiveKit might have different ways to access participants
          let participantsToProcess = [];
          
          // Method 1: Using room.participants Map (standard in newer versions)
          if (room.participants instanceof Map && room.participants.size > 0) {
            console.log('Debug: Found', room.participants.size, 'existing participants using Map');
            room.participants.forEach((participant, sid) => {
              participantsToProcess.push(participant);
            });
          } 
          // Method 2: Using getParticipants() method (used in some versions)
          else if (typeof room.getParticipants === 'function') {
            const participants = room.getParticipants();
            if (participants && participants.length > 0) {
              console.log('Debug: Found', participants.length, 'existing participants using getParticipants()');
              participantsToProcess = participants;
            }
          }
          // Method 3: Using room.remoteParticipants (used in some older versions)
          else if (room.remoteParticipants) {
            const remoteParticipants = room.remoteParticipants;
            if (remoteParticipants instanceof Map && remoteParticipants.size > 0) {
              console.log('Debug: Found', remoteParticipants.size, 'existing participants using remoteParticipants');
              remoteParticipants.forEach((participant, sid) => {
                participantsToProcess.push(participant);
              });
            }
          }
          
          // Process any participants we found
          if (participantsToProcess.length > 0) {
            console.log('Debug: Processing', participantsToProcess.length, 'participants');
            participantsToProcess.forEach(participant => {
              console.log('Debug: Processing existing participant:', participant.identity, 'SID:', participant.sid);
              handleParticipantConnected(participant);
            });
          } else {
            console.log('Debug: No existing participants found in the room using any method');
          }
        } catch (participantError) {
          console.error('Debug: Error handling existing participants:', participantError);
          console.error('Debug: Error stack:', participantError.stack);
        }
        
        console.log('Connected to room:', room.name);
      } catch (error) {
        console.error('Error connecting to room:', error);
        throw error;
      }
    }
    
    // Set up room event listeners
    function setupRoomEventListeners(room) {
      room.on(LivekitClient.RoomEvent.ParticipantConnected, handleParticipantConnected);
      room.on(LivekitClient.RoomEvent.ParticipantDisconnected, handleParticipantDisconnected);
      room.on(LivekitClient.RoomEvent.DataReceived, handleDataReceived);
      room.on(LivekitClient.RoomEvent.Disconnected, handleRoomDisconnected);
      room.on(LivekitClient.RoomEvent.Reconnected, handleRoomReconnected);
      room.on(LivekitClient.RoomEvent.Reconnecting, handleRoomReconnecting);
      room.on(LivekitClient.RoomEvent.LocalTrackPublished, handleLocalTrackPublished);
      room.on(LivekitClient.RoomEvent.LocalTrackUnpublished, handleLocalTrackUnpublished);
    }
    
    // Handle participant connected event
    function handleParticipantConnected(participant) {
      console.log('Debug: Participant connected:', participant.identity);
      console.log('Debug: Participant SID:', participant.sid);
      console.log('Debug: Participant object type:', typeof participant);
      console.log('Debug: Participant properties:');
      for (const prop in participant) {
        console.log(`Debug: Participant.${prop} type:`, typeof participant[prop]);
      }
      console.log('Debug: Participant has tracks:', participant.trackPublications ? participant.trackPublications.size : 'no trackPublications');
      
      // Log all track publications
      if (participant.trackPublications) {
        console.log('Debug: Track publications for', participant.identity, ':');
        participant.trackPublications.forEach((publication, trackSid) => {
          console.log(`Debug: Track ${trackSid}:`, publication.kind, 'isSubscribed:', publication.isSubscribed);
        });
      }
      
      renderRemoteParticipant(participant);
      
      // Set up participant event listeners
      console.log('Debug: Setting up event listeners for participant', participant.identity);
      participant.on(LivekitClient.ParticipantEvent.TrackSubscribed, (track, publication) => {
        console.log('Debug: TrackSubscribed event fired for', participant.identity);
        console.log('Debug: Track info:', track.kind, 'publication:', publication.trackSid);
        // Pass the participant from the outer scope, not from the event parameters
        handleTrackSubscribed(track, publication, participant);
      });
      
      participant.on(LivekitClient.ParticipantEvent.TrackUnsubscribed, (track, publication) => {
        console.log('Debug: TrackUnsubscribed event fired for', participant.identity);
        console.log('Debug: Track info:', track.kind, 'publication:', publication.trackSid);
        // Pass the participant from the outer scope, not from the event parameters
        handleTrackUnsubscribed(track, publication, participant);
      });
      
      // Listen for track published events
      participant.on(LivekitClient.ParticipantEvent.TrackPublished, (publication) => {
        console.log('Debug: TrackPublished event fired for', participant.identity, 'track:', publication.trackSid, 'kind:', publication.kind);
        console.log('Debug: Publication object:', publication);
        console.log('Debug: Publication properties:');
        for (const prop in publication) {
          if (typeof publication[prop] !== 'function') {
            console.log(`Debug: Publication.${prop}:`, publication[prop]);
          }
        }
        console.log('Debug: Publication methods:');
        for (const prop in publication) {
          if (typeof publication[prop] === 'function') {
            console.log(`Debug: Publication has method ${prop}`);
          }
        }
        
        // In LiveKit, we need to check if the track is already subscribed
        if (publication.isSubscribed && publication.track) {
          console.log('Debug: Track is already subscribed on publish, attaching');
          console.log('Debug: Track object:', publication.track);
          handleTrackSubscribed(publication.track, publication, participant);
        } else {
          console.log('Debug: Track published but not yet subscribed, waiting for subscription');
          console.log('Debug: isSubscribed:', publication.isSubscribed);
          console.log('Debug: track:', publication.track);
        }
      });
    }
    
    // Handle participant disconnected event
    function handleParticipantDisconnected(participant) {
      console.log('Participant disconnected:', participant.identity);
      const participantElement = document.getElementById(`participant-${participant.sid}`);
      if (participantElement) {
        participantElement.remove();
      }
    }
    
    // Handle data received event (for chat)
    function handleDataReceived(payload, participant) {
      if (!participant) return;
      
      try {
        console.log('Debug: Received data payload:', payload);
        console.log('Debug: Payload type:', typeof payload);
        
        // Safely decode the payload
        let decodedData;
        if (payload instanceof Uint8Array) {
          decodedData = new TextDecoder().decode(payload);
        } else if (typeof payload === 'string') {
          decodedData = payload;
        } else {
          console.error('Debug: Unexpected payload type:', typeof payload);
          return;
        }
        
        console.log('Debug: Decoded data:', decodedData);
        
        // Handle empty or invalid data
        if (!decodedData || decodedData.trim() === '') {
          console.error('Debug: Empty data received');
          return;
        }
        
        // Parse the JSON data
        const data = JSON.parse(decodedData);
        console.log('Debug: Parsed data:', data);
        
        if (data.type === 'chat') {
          addChatMessage(participant.identity, data.message);
        }
      } catch (error) {
        console.error('Error parsing data:', error);
        console.error('Debug: Error details:', error.message);
        console.error('Debug: Error stack:', error.stack);
      }
    }
    
    // Handle room disconnected event
    function handleRoomDisconnected() {
      console.log('Disconnected from room');
      window.location.href = '/';
    }
    
    // Handle room reconnected event
    function handleRoomReconnected() {
      console.log('Reconnected to room');
    }
    
    // Handle room reconnecting event
    function handleRoomReconnecting() {
      console.log('Reconnecting to room...');
    }
    
    // Handle local track published event
    function handleLocalTrackPublished(track, publication) {
      console.log('Local track published:', track.kind);
      updateLocalParticipantUI();
    }
    
    // Handle local track unpublished event
    function handleLocalTrackUnpublished(track, publication) {
      console.log('Local track unpublished:', track.kind);
      updateLocalParticipantUI();
    }
    
    // Handle track subscribed event
    function handleTrackSubscribed(track, publication, participant) {
      console.log('Debug: handleTrackSubscribed called with track:', track ? track.kind : 'undefined');
      console.log('Debug: publication:', publication ? publication.trackSid : 'undefined');
      console.log('Debug: participant:', participant ? participant.identity : 'undefined');
      
      // Check if participant is defined
      if (!participant) {
        console.error('Debug: participant is undefined in handleTrackSubscribed');
        return;
      }
      
      // Check if track is defined
      if (!track) {
        console.error('Debug: track is undefined in handleTrackSubscribed');
        return;
      }
      
      console.log('Track subscribed:', track.kind, 'from', participant.identity);
      
      try {
        // Get or create participant element
        let participantElement = document.getElementById(`participant-${participant.sid}`);
        console.log('Debug: participantElement found:', !!participantElement);
        
        if (!participantElement) {
          console.log('Debug: Creating participant element for', participant.identity);
          // If the participant element doesn't exist yet, create it
          renderRemoteParticipant(participant);
          // Try to get the element again
          participantElement = document.getElementById(`participant-${participant.sid}`);
          if (!participantElement) {
            console.error('Debug: Failed to create participant element');
            return;
          }
        }
        
        const videoContainer = participantElement.querySelector('.video-container');
        if (!videoContainer) {
          console.error('Debug: Video container not found for participant', participant.identity);
          return;
        }
        
        if (track.kind === 'video') {
          console.log('Debug: Attaching video track', publication.trackSid);
          
          // Remove any placeholder
          const placeholder = videoContainer.querySelector('.video-placeholder');
          if (placeholder) {
            placeholder.remove();
          }
          
          // Create a new video element
          const videoElement = document.createElement('video');
          videoElement.id = `video-${publication.trackSid}`;
          videoElement.autoplay = true;
          videoElement.playsInline = true;
          videoElement.muted = false; // Ensure remote videos are not muted
          
          // Clear existing content and append the new video element
          videoContainer.innerHTML = '';
          videoContainer.appendChild(videoElement);
          
          // Attach the track to the video element
          console.log('Debug: About to attach video track to element');
          track.attach(videoElement);
          console.log('Debug: Video track attached successfully');
        } else if (track.kind === 'audio') {
          console.log('Debug: Attaching audio track', publication.trackSid);
          
          // Create a new audio element
          const audioElement = document.createElement('audio');
          audioElement.id = `audio-${publication.trackSid}`;
          audioElement.autoplay = true;
          audioElement.controls = false;
          
          // Append audio element to the body to ensure it plays
          document.body.appendChild(audioElement);
          
          // Attach the track to the audio element
          track.attach(audioElement);
          console.log('Debug: Audio track attached successfully');
        }
      } catch (error) {
        console.error('Debug: Error in handleTrackSubscribed:', error);
      }
    }
    
    // Handle track unsubscribed event
    function handleTrackUnsubscribed(track, publication, participant) {
      console.log('Debug: handleTrackUnsubscribed called with track:', track ? track.kind : 'undefined');
      console.log('Debug: publication:', publication ? publication.trackSid : 'undefined');
      console.log('Debug: participant:', participant ? participant.identity : 'undefined');
      
      // Check if participant is defined
      if (!participant) {
        console.error('Debug: participant is undefined in handleTrackUnsubscribed');
        return;
      }
      
      console.log('Track unsubscribed:', track.kind, 'from', participant.identity);
      
      if (track.kind === 'video') {
        console.log('Debug: Detaching video track', publication.trackSid);
        const videoElement = document.getElementById(`video-${publication.trackSid}`);
        console.log('Debug: videoElement found:', !!videoElement);
        if (videoElement) {
          try {
            track.detach(videoElement);
            videoElement.remove();
            console.log('Debug: Video element detached and removed');
          } catch (detachError) {
            console.error('Debug: Error detaching video track:', detachError);
          }
        }
      } else if (track.kind === 'audio') {
        console.log('Debug: Detaching audio track', publication.trackSid);
        const audioElement = document.getElementById(`audio-${publication.trackSid}`);
        console.log('Debug: audioElement found:', !!audioElement);
        if (audioElement) {
          try {
            track.detach(audioElement);
            audioElement.remove();
            console.log('Debug: Audio element detached and removed');
          } catch (detachError) {
            console.error('Debug: Error detaching audio track:', detachError);
          }
        }
      }
    }
    
    // Render local participant
    function renderLocalParticipant() {
      console.log('Debug: renderLocalParticipant called');
      updateLocalParticipantUI();
    }
    
    // Update local participant UI
    function updateLocalParticipantUI() {
      console.log('Debug: updateLocalParticipantUI called');
      if (!localParticipant) {
        console.log('Debug: localParticipant is null or undefined, returning');
        return;
      }
      
      console.log('Debug: Looking for local-participant video container');
      const videoContainer = document.querySelector('#local-participant .video-container');
      console.log('Debug: videoContainer found:', !!videoContainer);
      if (!videoContainer) {
        console.error('Debug: videoContainer not found in DOM');
        console.log('Debug: DOM structure:', document.body.innerHTML);
        return;
      }
      videoContainer.innerHTML = '';
      
      console.log('Debug: Getting camera track publication');
      const cameraPublication = localParticipant.getTrackPublication('camera');
      console.log('Debug: cameraPublication:', cameraPublication ? 'found' : 'not found');
      
      // For local tracks, we don't check isSubscribed since local tracks are published, not subscribed
      if (cameraPublication && cameraPublication.track) {
        console.log('Debug: Camera track found, attaching to video element');
        console.log('Debug: localVideo exists:', !!localVideo);
        try {
          // Create a new video element if it doesn't exist
          if (!localVideo) {
            console.log('Debug: Creating new video element for local video');
            localVideo = document.createElement('video');
            localVideo.id = 'local-video';
            localVideo.autoplay = true;
            localVideo.muted = true;
            localVideo.playsInline = true;
          }
          
          cameraPublication.track.attach(localVideo);
          videoContainer.appendChild(localVideo);
          console.log('Debug: Video attached and appended successfully');
        } catch (attachError) {
          console.error('Debug: Error attaching video:', attachError);
        }
      } else {
        console.log('Debug: Creating placeholder for camera');
        try {
          const placeholderDiv = document.createElement('div');
          placeholderDiv.className = 'video-placeholder';
          placeholderDiv.textContent = '📷';
          videoContainer.appendChild(placeholderDiv);
          console.log('Debug: Placeholder created and appended successfully');
        } catch (placeholderError) {
          console.error('Debug: Error creating placeholder:', placeholderError);
        }
      }
      
      // Update control buttons
      console.log('Debug: Updating control buttons');
      try {
        console.log('Debug: toggleVideoBtn exists:', !!toggleVideoBtn);
        console.log('Debug: toggleAudioBtn exists:', !!toggleAudioBtn);
        if (toggleVideoBtn) {
          toggleVideoBtn.classList.toggle('active', localParticipant.isCameraEnabled);
        } else {
          console.error('Debug: toggleVideoBtn not found');
        }
        
        if (toggleAudioBtn) {
          toggleAudioBtn.classList.toggle('active', localParticipant.isMicrophoneEnabled);
        } else {
          console.error('Debug: toggleAudioBtn not found');
        }
        console.log('Debug: Control buttons updated successfully');
      } catch (buttonError) {
        console.error('Debug: Error updating control buttons:', buttonError);
      }
    }
    
    // Render remote participant
    function renderRemoteParticipant(participant) {
      console.log('Debug: renderRemoteParticipant called for participant:', participant.identity, 'SID:', participant.sid);
      
      // Check if participant element already exists
      const existingElement = document.getElementById(`participant-${participant.sid}`);
      if (existingElement) {
        console.log('Debug: Participant element already exists, skipping creation');
        return;
      }
      
      try {
        // Get the container for remote participants
        const remoteParticipants = document.getElementById('remote-participants');
        if (!remoteParticipants) {
          console.error('Debug: remote-participants container not found in DOM');
          console.log('Debug: DOM structure:', document.body.innerHTML);
          return;
        }
        
        // Create participant element
        const participantElement = document.createElement('div');
        participantElement.id = `participant-${participant.sid}`;
        participantElement.className = 'participant-tile';
        
        // Create video container
        const videoContainer = document.createElement('div');
        videoContainer.className = 'video-container';
        
        // Add placeholder until video arrives
        const placeholderDiv = document.createElement('div');
        placeholderDiv.className = 'video-placeholder';
        placeholderDiv.textContent = '📷';
        videoContainer.appendChild(placeholderDiv);
        
        // Create participant info
        const participantInfo = document.createElement('div');
        participantInfo.className = 'participant-info';
        
        const participantName = document.createElement('div');
        participantName.className = 'participant-name';
        participantName.textContent = participant.identity;
        
        participantInfo.appendChild(participantName);
        
        // Assemble the participant tile
        participantElement.appendChild(videoContainer);
        participantElement.appendChild(participantInfo);
        
        // Add to the DOM
        remoteParticipants.appendChild(participantElement);
        console.log('Debug: Added participant element to DOM');
        
        // Attach existing tracks
        console.log('Debug: Checking for existing tracks for participant:', participant.identity);
        console.log('Debug: participant.trackPublications:', participant.trackPublications);
      
      // In LiveKit, tracks are stored in a Map called trackPublications
      if (participant.trackPublications && participant.trackPublications.size > 0) {
        console.log('Debug: Found', participant.trackPublications.size, 'track publications');
        
        // Convert the Map to an array and iterate
        Array.from(participant.trackPublications.values()).forEach(publication => {
          console.log('Debug: Processing track publication:', publication.trackSid, 'kind:', publication.kind, 'isSubscribed:', publication.isSubscribed);
          
          // Check if the track is already subscribed and has a track object
          if (publication.isSubscribed && publication.track) {
            console.log('Debug: Track is already subscribed, attaching');
            // Use a timeout to ensure the DOM element is fully created
            setTimeout(() => {
              handleTrackSubscribed(publication.track, publication, participant);
            }, 100);
          } else {
            console.log('Debug: Track exists but is not yet subscribed');
          }
        });
      } else {
        console.log('Debug: No track publications found for participant');
      }
      } catch (error) {
        console.error('Debug: Error in renderRemoteParticipant:', error);
      }
    }
    
    // Control button event listeners
    toggleVideoBtn.addEventListener('click', async () => {
      if (!localParticipant) return;
      
      console.log('Debug: Toggle video clicked, current state:', localParticipant.isCameraEnabled);
      
      try {
        // Detect Firefox browser
        const isFirefox = navigator.userAgent.toLowerCase().indexOf('firefox') > -1;
        
        if (localParticipant.isCameraEnabled) {
          console.log('Debug: Disabling camera');
          await localParticipant.setCameraEnabled(false);
          
          // Firefox-specific handling for video disabling
          if (isFirefox) {
            console.log('Debug: Firefox-specific video disable handling');
            // Get the local video container
            const videoContainer = document.querySelector('#local-participant .video-container');
            if (videoContainer) {
              // Clear existing content
              videoContainer.innerHTML = '';
              
              // Add a placeholder
              const placeholderDiv = document.createElement('div');
              placeholderDiv.className = 'video-placeholder';
              placeholderDiv.textContent = '📷';
              videoContainer.appendChild(placeholderDiv);
              
              console.log('Debug: Added placeholder for disabled camera in Firefox');
            }
          }
        } else {
          console.log('Debug: Enabling camera');
          await localParticipant.setCameraEnabled(true);
          
          // Update UI after enabling camera
          setTimeout(() => {
            updateLocalParticipantUI();
          }, 100);
        }
        
        console.log('Debug: Camera state after toggle:', localParticipant.isCameraEnabled);
        toggleVideoBtn.classList.toggle('active', localParticipant.isCameraEnabled);
      } catch (error) {
        console.error('Debug: Error toggling video:', error);
      }
    });
    
    toggleAudioBtn.addEventListener('click', async () => {
      if (!localParticipant) return;
      
      if (localParticipant.isMicrophoneEnabled) {
        await localParticipant.setMicrophoneEnabled(false);
      } else {
        await localParticipant.setMicrophoneEnabled(true);
      }
      
      toggleAudioBtn.classList.toggle('active', localParticipant.isMicrophoneEnabled);
    });
    
    shareScreenBtn.addEventListener('click', async () => {
      if (!localParticipant) return;
      
      try {
        if (shareScreenBtn.classList.contains('active')) {
          await localParticipant.setScreenShareEnabled(false);
        } else {
          await localParticipant.setScreenShareEnabled(true);
        }
        
        shareScreenBtn.classList.toggle('active');
      } catch (error) {
        console.error('Error toggling screen share:', error);
        alert(`Error toggling screen share: ${error.message}`);
      }
    });
    
    toggleChatBtn.addEventListener('click', () => {
      chatPanel.style.display = chatPanel.style.display === 'none' ? 'flex' : 'none';
      toggleChatBtn.classList.toggle('active', chatPanel.style.display === 'flex');
    });
    
    closeChatBtn.addEventListener('click', () => {
      chatPanel.style.display = 'none';
      toggleChatBtn.classList.remove('active');
    });
    
    leaveRoomBtn.addEventListener('click', async () => {
      if (room) {
        await room.disconnect();
      }
      window.location.href = '/';
    });
    
    // Chat form event listener
    chatForm.addEventListener('submit', (event) => {
      event.preventDefault();
      
      const message = chatMessageInput.value.trim();
      if (!message || !room || !localParticipant) return;
      
      // Send message to all participants
      const data = {
        type: 'chat',
        message
      };
      
      try {
        console.log('Debug: Sending chat data:', data);
        const jsonString = JSON.stringify(data);
        console.log('Debug: JSON string to send:', jsonString);
        
        // Convert string to Uint8Array for LiveKit
        const encoder = new TextEncoder();
        const dataToSend = encoder.encode(jsonString);
        
        // Publish data to all participants
        room.localParticipant.publishData(dataToSend, LivekitClient.DataPacket_Kind.RELIABLE);
        console.log('Debug: Chat message sent successfully');
      } catch (error) {
        console.error('Debug: Error sending chat message:', error);
      }
      
      // Add message to chat panel
      addChatMessage('You', message, true);
      
      // Clear input
      chatMessageInput.value = '';
    });
  });
  