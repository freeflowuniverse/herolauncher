package videoconf

import "errors"

// Custom errors for the videoconf package
var (
	ErrMissingLiveKitConfig = errors.New("missing LiveKit configuration: LIVEKIT_URL, LIVEKIT_API_KEY, and LIVEKIT_API_SECRET must be set")
	ErrRoomNotFound         = errors.New("room not found")
	ErrInvalidRoomName      = errors.New("invalid room name")
	ErrInvalidParticipant   = errors.New("invalid participant name")
)
