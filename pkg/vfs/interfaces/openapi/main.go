package openapi

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolib/lib/vfs"
)

// StartVFSOpenAPIServer starts an OpenAPI server for the given VFS implementation
func StartVFSOpenAPIServer(vfsImpl vfs.VFSImplementation, port int) error {
	server := NewServer(vfsImpl, port)
	return server.Start()
}
