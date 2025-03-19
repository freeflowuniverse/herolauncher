package openapi

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
)

// handleGet handles GET requests (read operations)
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request, path string) {
	// Check if path exists
	if !s.vfsImpl.exists(path) {
		http.Error(w, "File or directory not found", http.StatusNotFound)
		return
	}

	// Get the entry
	entry, err := s.vfsImpl.get(path)
	if err != nil {
		http.Error(w, "Error getting entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle based on entry type
	if entry.is_dir() {
		// List directory contents
		entries, err := s.vfsImpl.dir_list(path)
		if err != nil {
			http.Error(w, "Error listing directory: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert entries to JSON-friendly format
		type EntryInfo struct {
			Name     string `json:"name"`
			IsDir    bool   `json:"is_dir"`
			IsFile   bool   `json:"is_file"`
			IsSymlink bool  `json:"is_symlink"`
		}

		result := make([]EntryInfo, 0, len(entries))
		for _, e := range entries {
			entryPath, err := s.vfsImpl.get_path(&e)
			if err != nil {
				continue
			}
			
			result = append(result, EntryInfo{
				Name:     filepath.Base(entryPath),
				IsDir:    e.is_dir(),
				IsFile:   e.is_file(),
				IsSymlink: e.is_symlink(),
			})
		}

		// Return as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else if entry.is_file() {
		// Read file contents
		data, err := s.vfsImpl.file_read(path)
		if err != nil {
			http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return file contents
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(data)
	} else if entry.is_symlink() {
		// Read symlink target
		target, err := s.vfsImpl.link_read(path)
		if err != nil {
			http.Error(w, "Error reading symlink: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return target as plain text
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(target))
	} else {
		http.Error(w, "Unknown entry type", http.StatusInternalServerError)
	}
}

// handlePut handles PUT requests (create/update operations)
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request, path string) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the path exists
	exists := s.vfsImpl.exists(path)

	// Handle file creation/update
	if exists {
		// Update existing file
		err = s.vfsImpl.file_write(path, body)
		if err != nil {
			http.Error(w, "Error updating file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("File updated successfully"))
	} else {
		// Create new file
		_, err = s.vfsImpl.file_create(path)
		if err != nil {
			http.Error(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Write content to the new file
		err = s.vfsImpl.file_write(path, body)
		if err != nil {
			http.Error(w, "Error writing to file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File created successfully"))
	}
}

// handlePost handles POST requests (create directories, append to files, create symlinks)
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request, path string) {
	// Parse the operation type from query parameters
	operation := r.URL.Query().Get("op")

	switch operation {
	case "mkdir":
		// Create directory
		_, err := s.vfsImpl.dir_create(path)
		if err != nil {
			http.Error(w, "Error creating directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Directory created successfully"))

	case "append":
		// Append to file
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = s.vfsImpl.file_concatenate(path, body)
		if err != nil {
			http.Error(w, "Error appending to file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Content appended successfully"))

	case "symlink":
		// Create symlink
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Missing target parameter for symlink", http.StatusBadRequest)
			return
		}

		_, err := s.vfsImpl.link_create(target, path)
		if err != nil {
			http.Error(w, "Error creating symlink: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Symlink created successfully"))

	case "copy":
		// Copy file or directory
		src := r.URL.Query().Get("src")
		if src == "" {
			http.Error(w, "Missing src parameter for copy operation", http.StatusBadRequest)
			return
		}

		_, err := s.vfsImpl.copy(src, path)
		if err != nil {
			http.Error(w, "Error copying: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Copy completed successfully"))

	case "move":
		// Move file or directory
		src := r.URL.Query().Get("src")
		if src == "" {
			http.Error(w, "Missing src parameter for move operation", http.StatusBadRequest)
			return
		}

		_, err := s.vfsImpl.move(src, path)
		if err != nil {
			http.Error(w, "Error moving: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Move completed successfully"))

	case "rename":
		// Rename file or directory
		oldPath := r.URL.Query().Get("old")
		if oldPath == "" {
			http.Error(w, "Missing old parameter for rename operation", http.StatusBadRequest)
			return
		}

		_, err := s.vfsImpl.rename(oldPath, path)
		if err != nil {
			http.Error(w, "Error renaming: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Rename completed successfully"))

	default:
		http.Error(w, "Unknown operation", http.StatusBadRequest)
	}
}

// handleDelete handles DELETE requests (delete operations)
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request, path string) {
	// Check if path exists
	if !s.vfsImpl.exists(path) {
		http.Error(w, "File or directory not found", http.StatusNotFound)
		return
	}

	// Get the entry to determine its type
	entry, err := s.vfsImpl.get(path)
	if err != nil {
		http.Error(w, "Error getting entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete based on entry type
	if entry.is_dir() {
		err = s.vfsImpl.dir_delete(path)
	} else if entry.is_file() {
		err = s.vfsImpl.file_delete(path)
	} else if entry.is_symlink() {
		err = s.vfsImpl.link_delete(path)
	} else {
		// Generic delete as fallback
		err = s.vfsImpl.delete(path)
	}

	if err != nil {
		http.Error(w, "Error deleting entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Entry deleted successfully"))
}
