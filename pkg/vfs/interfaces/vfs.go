package interfaces

// VFSManager defines the interface for virtual file system operations
type VFSManager interface {
	// UploadFile uploads a file to the virtual file system
	UploadFile(vfspath, dedupepath, filepath string) (UploadResult, error)

	// UploadDir uploads a directory to the virtual file system
	UploadDir(vfspath, dedupepath, dirpath string) (UploadDirResult, error)

	// DownloadFile downloads a file from the virtual file system
	DownloadFile(vfspath, dedupepath, destpath string) (DownloadResult, error)

	// ExportMeta exports metadata from the virtual file system
	ExportMeta(vfspath, destpath string) (ExportMetaResult, error)

	// ImportMeta imports metadata to the virtual file system
	ImportMeta(vfspath, sourcepath string) (ImportMetaResult, error)

	// ExportDedupe exports dedupe information from the virtual file system
	ExportDedupe(vfspath, dedupepath, destpath string) (ExportDedupeResult, error)

	// ImportDedupe imports dedupe information to the virtual file system
	ImportDedupe(vfspath, dedupepath, sourcepath string) (ImportDedupeResult, error)

	// Send sends files based on dedupe hashes to a destination
	Send(dedupepath, pubkeydest string, hashlist []string, secret string) (SendResult, error)

	// SendExist checks which dedupe hashes exist and returns a list
	SendExist(dedupepath, pubkeydest string, hashlist []string, secret string) (SendExistResult, error)

	// ExposeWebDAV exposes the virtual file system via WebDAV
	ExposeWebDAV(vfspath string, port int, username, password string) (ExposeWebDAVResult, error)

	// Expose9P exposes the virtual file system via 9P protocol
	Expose9P(vfspath string, port int, readonly bool) (Expose9PResult, error)
}

// UploadResult represents the result of a file upload operation
type UploadResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Hash    string `json:"hash"`
}

// UploadDirResult represents the result of a directory upload operation
type UploadDirResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	FilesProcessed int   `json:"files_processed"`
	TotalSize     int64  `json:"total_size"`
}

// DownloadResult represents the result of a file download operation
type DownloadResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Size    int64  `json:"size"`
}

// ExportMetaResult represents the result of a metadata export operation
type ExportMetaResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Entries int    `json:"entries"`
}

// ImportMetaResult represents the result of a metadata import operation
type ImportMetaResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	EntriesImported int   `json:"entries_imported"`
}

// ExportDedupeResult represents the result of a dedupe export operation
type ExportDedupeResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	HashesExported int   `json:"hashes_exported"`
}

// ImportDedupeResult represents the result of a dedupe import operation
type ImportDedupeResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	HashesImported int   `json:"hashes_imported"`
}

// SendResult represents the result of a send operation
type SendResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	HashesSent int   `json:"hashes_sent"`
	TotalSize int64  `json:"total_size"`
}

// SendExistResult represents the result of a send_exist operation
type SendExistResult struct {
	Success       bool     `json:"success"`
	Message       string   `json:"message"`
	ExistingHashes []string `json:"existing_hashes"`
}

// ExposeWebDAVResult represents the result of a WebDAV exposure operation
type ExposeWebDAVResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

// Expose9PResult represents the result of a 9P exposure operation
type Expose9PResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Address string `json:"address"`
}
