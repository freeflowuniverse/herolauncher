package srtrecorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestSRTRecorder(t *testing.T) {
	// Skip if ffmpeg is not installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	// Create a temporary directory for recordings
	tempDir, err := os.MkdirTemp("", "srtrecorder-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create SRT recorder
	port := 8090
	recorder := NewSRTRecorder(port, tempDir)

	// Start the recorder
	if err := recorder.Start(); err != nil {
		t.Fatalf("Failed to start SRT recorder: %v", err)
	}
	defer recorder.Stop()

	// Wait for recorder to fully start
	time.Sleep(1 * time.Second)

	// Generate a test video file using ffmpeg
	testVideoPath := filepath.Join(tempDir, "test_video.mp4")
	genCmd := exec.Command(
		"ffmpeg",
		"-f", "lavfi",
		"-i", "testsrc=duration=5:size=640x480:rate=30",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		testVideoPath,
	)
	
	if err := genCmd.Run(); err != nil {
		t.Fatalf("Failed to generate test video: %v", err)
	}

	// Start a goroutine to stream the test video to SRT
	streamDone := make(chan struct{})
	go func() {
		defer close(streamDone)
		
		// Stream the test video to SRT using ffmpeg
		streamCmd := exec.Command(
			"ffmpeg",
			"-re",
			"-i", testVideoPath,
			"-c", "copy",
			"-f", "mpegts",
			fmt.Sprintf("srt://127.0.0.1:%d?mode=caller", port),
		)
		
		// Capture output for debugging
		streamCmd.Stdout = os.Stdout
		streamCmd.Stderr = os.Stderr
		
		if err := streamCmd.Run(); err != nil {
			t.Logf("FFmpeg streaming command failed: %v", err)
		}
	}()

	// Wait for streaming to complete (with timeout)
	select {
	case <-streamDone:
		// Streaming completed normally
	case <-time.After(10 * time.Second):
		t.Log("Streaming timeout reached, continuing with test")
	}

	// Wait a bit for the recorder to finish writing
	time.Sleep(2 * time.Second)

	// Stop the recorder
	if err := recorder.Stop(); err != nil {
		t.Fatalf("Failed to stop SRT recorder: %v", err)
	}

	// Check if at least one recording file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	recordingFound := false
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".ts" {
			recordingFound = true
			
			// Verify the file has content
			info, err := file.Info()
			if err != nil {
				t.Fatalf("Failed to get file info: %v", err)
			}
			
			if info.Size() == 0 {
				t.Errorf("Recording file is empty: %s", file.Name())
			} else {
				t.Logf("Found valid recording: %s (%d bytes)", file.Name(), info.Size())
			}
		}
	}

	if !recordingFound {
		t.Errorf("No recording files were created")
	}
}
