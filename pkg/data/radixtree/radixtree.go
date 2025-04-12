// Package radixtree provides a persistent radix tree implementation using the ourdb package for storage
package radixtree

import (
	"errors"

	"github.com/freeflowuniverse/herolauncher/pkg/ourdb"
)

// Node represents a node in the radix tree
type Node struct {
	KeySegment string    // The segment of the key stored at this node
	Value      []byte    // Value stored at this node (empty if not a leaf)
	Children   []NodeRef // References to child nodes
	IsLeaf     bool      // Whether this node is a leaf node
}

// NodeRef is a reference to a node in the database
type NodeRef struct {
	KeyPart string // The key segment for this child
	NodeID  uint32 // Database ID of the node
}

// RadixTree represents a radix tree data structure
type RadixTree struct {
	DB     *ourdb.OurDB // Database for persistent storage
	RootID uint32       // Database ID of the root node
}

// NewArgs contains arguments for creating a new RadixTree
type NewArgs struct {
	Path  string // Path to the database
	Reset bool   // Whether to reset the database
}

// New creates a new radix tree with the specified database path
func New(args NewArgs) (*RadixTree, error) {
	config := ourdb.DefaultConfig()
	config.Path = args.Path
	config.RecordSizeMax = 1024 * 4 // 4KB max record size
	config.IncrementalMode = true
	config.Reset = args.Reset

	db, err := ourdb.New(config)
	if err != nil {
		return nil, err
	}

	var rootID uint32 = 1 // First ID in ourdb is 1
	nextID, err := db.GetNextID()
	if err != nil {
		return nil, err
	}

	if nextID == 1 {
		// Create new root node
		root := Node{
			KeySegment: "",
			Value:      []byte{},
			Children:   []NodeRef{},
			IsLeaf:     false,
		}
		rootData := serializeNode(root)
		rootID, err = db.Set(ourdb.OurDBSetArgs{
			Data: rootData,
		})
		if err != nil {
			return nil, err
		}
		if rootID != 1 {
			return nil, errors.New("expected root ID to be 1")
		}
	} else {
		// Use existing root node
		_, err := db.Get(1) // Verify root node exists
		if err != nil {
			return nil, err
		}
	}

	return &RadixTree{
		DB:     db,
		RootID: rootID,
	}, nil
}

// Set sets a key-value pair in the tree
func (rt *RadixTree) Set(key string, value []byte) error {
	currentID := rt.RootID
	offset := 0

	// Handle empty key case
	if len(key) == 0 {
		rootData, err := rt.DB.Get(currentID)
		if err != nil {
			return err
		}
		rootNode, err := deserializeNode(rootData)
		if err != nil {
			return err
		}
		rootNode.IsLeaf = true
		rootNode.Value = value
		_, err = rt.DB.Set(ourdb.OurDBSetArgs{
			ID:   &currentID,
			Data: serializeNode(rootNode),
		})
		return err
	}

	for offset < len(key) {
		nodeData, err := rt.DB.Get(currentID)
		if err != nil {
			return err
		}
		node, err := deserializeNode(nodeData)
		if err != nil {
			return err
		}

		// Find matching child
		matchedChild := -1
		for i, child := range node.Children {
			if hasPrefix(key[offset:], child.KeyPart) {
				matchedChild = i
				break
			}
		}

		if matchedChild == -1 {
			// No matching child found, create new leaf node
			keyPart := key[offset:]
			newNode := Node{
				KeySegment: keyPart,
				Value:      value,
				Children:   []NodeRef{},
				IsLeaf:     true,
			}
			newID, err := rt.DB.Set(ourdb.OurDBSetArgs{
				Data: serializeNode(newNode),
			})
			if err != nil {
				return err
			}

			// Create new child reference and update parent node
			node.Children = append(node.Children, NodeRef{
				KeyPart: keyPart,
				NodeID:  newID,
			})

			// Update parent node in DB
			_, err = rt.DB.Set(ourdb.OurDBSetArgs{
				ID:   &currentID,
				Data: serializeNode(node),
			})
			return err
		}

		child := node.Children[matchedChild]
		commonPrefix := getCommonPrefix(key[offset:], child.KeyPart)

		if len(commonPrefix) < len(child.KeyPart) {
			// Split existing node
			childData, err := rt.DB.Get(child.NodeID)
			if err != nil {
				return err
			}
			childNode, err := deserializeNode(childData)
			if err != nil {
				return err
			}

			// Create new intermediate node
			newNode := Node{
				KeySegment: child.KeyPart[len(commonPrefix):],
				Value:      childNode.Value,
				Children:   childNode.Children,
				IsLeaf:     childNode.IsLeaf,
			}
			newID, err := rt.DB.Set(ourdb.OurDBSetArgs{
				Data: serializeNode(newNode),
			})
			if err != nil {
				return err
			}

			// Update current node
			node.Children[matchedChild] = NodeRef{
				KeyPart: commonPrefix,
				NodeID:  newID,
			}
			_, err = rt.DB.Set(ourdb.OurDBSetArgs{
				ID:   &currentID,
				Data: serializeNode(node),
			})
			if err != nil {
				return err
			}
		}

		if offset+len(commonPrefix) == len(key) {
			// Update value at existing node
			childData, err := rt.DB.Get(child.NodeID)
			if err != nil {
				return err
			}
			childNode, err := deserializeNode(childData)
			if err != nil {
				return err
			}
			childNode.Value = value
			childNode.IsLeaf = true
			_, err = rt.DB.Set(ourdb.OurDBSetArgs{
				ID:   &child.NodeID,
				Data: serializeNode(childNode),
			})
			return err
		}

		offset += len(commonPrefix)
		currentID = child.NodeID
	}

	return nil
}

// Get retrieves a value by key from the tree
func (rt *RadixTree) Get(key string) ([]byte, error) {
	currentID := rt.RootID
	offset := 0

	// Handle empty key case
	if len(key) == 0 {
		rootData, err := rt.DB.Get(currentID)
		if err != nil {
			return nil, err
		}
		rootNode, err := deserializeNode(rootData)
		if err != nil {
			return nil, err
		}
		if rootNode.IsLeaf {
			return rootNode.Value, nil
		}
		return nil, errors.New("key not found")
	}

	for offset < len(key) {
		nodeData, err := rt.DB.Get(currentID)
		if err != nil {
			return nil, err
		}
		node, err := deserializeNode(nodeData)
		if err != nil {
			return nil, err
		}

		found := false
		for _, child := range node.Children {
			if hasPrefix(key[offset:], child.KeyPart) {
				if offset+len(child.KeyPart) == len(key) {
					childData, err := rt.DB.Get(child.NodeID)
					if err != nil {
						return nil, err
					}
					childNode, err := deserializeNode(childData)
					if err != nil {
						return nil, err
					}
					if childNode.IsLeaf {
						return childNode.Value, nil
					}
				}
				currentID = child.NodeID
				offset += len(child.KeyPart)
				found = true
				break
			}
		}

		if !found {
			return nil, errors.New("key not found")
		}
	}

	return nil, errors.New("key not found")
}

// Update updates the value at a given key prefix, preserving the prefix while replacing the remainder
func (rt *RadixTree) Update(prefix string, newValue []byte) error {
	currentID := rt.RootID
	offset := 0

	// Handle empty prefix case
	if len(prefix) == 0 {
		return errors.New("empty prefix not allowed")
	}

	for offset < len(prefix) {
		nodeData, err := rt.DB.Get(currentID)
		if err != nil {
			return err
		}
		node, err := deserializeNode(nodeData)
		if err != nil {
			return err
		}

		found := false
		for _, child := range node.Children {
			if hasPrefix(prefix[offset:], child.KeyPart) {
				if offset+len(child.KeyPart) == len(prefix) {
					// Found exact prefix match
					childData, err := rt.DB.Get(child.NodeID)
					if err != nil {
						return err
					}
					childNode, err := deserializeNode(childData)
					if err != nil {
						return err
					}
					if childNode.IsLeaf {
						// Update the value
						childNode.Value = newValue
						_, err = rt.DB.Set(ourdb.OurDBSetArgs{
							ID:   &child.NodeID,
							Data: serializeNode(childNode),
						})
						return err
					}
				}
				currentID = child.NodeID
				offset += len(child.KeyPart)
				found = true
				break
			}
		}

		if !found {
			return errors.New("prefix not found")
		}
	}

	return errors.New("prefix not found")
}

// Delete deletes a key from the tree
func (rt *RadixTree) Delete(key string) error {
	currentID := rt.RootID
	offset := 0
	var path []NodeRef

	// Find the node to delete
	for offset < len(key) {
		nodeData, err := rt.DB.Get(currentID)
		if err != nil {
			return err
		}
		node, err := deserializeNode(nodeData)
		if err != nil {
			return err
		}

		found := false
		for _, child := range node.Children {
			if hasPrefix(key[offset:], child.KeyPart) {
				path = append(path, child)
				currentID = child.NodeID
				offset += len(child.KeyPart)
				found = true

				// Check if we've matched the full key
				if offset == len(key) {
					childData, err := rt.DB.Get(child.NodeID)
					if err != nil {
						return err
					}
					childNode, err := deserializeNode(childData)
					if err != nil {
						return err
					}
					if childNode.IsLeaf {
						found = true
						break
					}
				}
				break
			}
		}

		if !found {
			return errors.New("key not found")
		}
	}

	if len(path) == 0 {
		return errors.New("key not found")
	}

	// Get the node to delete
	lastNodeID := path[len(path)-1].NodeID
	lastNodeData, err := rt.DB.Get(lastNodeID)
	if err != nil {
		return err
	}
	lastNode, err := deserializeNode(lastNodeData)
	if err != nil {
		return err
	}

	// If the node has children, just mark it as non-leaf
	if len(lastNode.Children) > 0 {
		lastNode.IsLeaf = false
		lastNode.Value = []byte{}
		_, err = rt.DB.Set(ourdb.OurDBSetArgs{
			ID:   &lastNodeID,
			Data: serializeNode(lastNode),
		})
		return err
	}

	// If node has no children, remove it from parent
	if len(path) > 1 {
		parentNodeID := path[len(path)-2].NodeID
		parentNodeData, err := rt.DB.Get(parentNodeID)
		if err != nil {
			return err
		}
		parentNode, err := deserializeNode(parentNodeData)
		if err != nil {
			return err
		}

		// Remove child from parent
		for i, child := range parentNode.Children {
			if child.NodeID == lastNodeID {
				// Remove child at index i
				parentNode.Children = append(parentNode.Children[:i], parentNode.Children[i+1:]...)
				break
			}
		}

		_, err = rt.DB.Set(ourdb.OurDBSetArgs{
			ID:   &parentNodeID,
			Data: serializeNode(parentNode),
		})
		if err != nil {
			return err
		}

		// Delete the node from the database
		return rt.DB.Delete(lastNodeID)
	} else {
		// If this is a direct child of the root, just mark it as non-leaf
		lastNode.IsLeaf = false
		lastNode.Value = []byte{}
		_, err = rt.DB.Set(ourdb.OurDBSetArgs{
			ID:   &lastNodeID,
			Data: serializeNode(lastNode),
		})
		return err
	}
}

// List lists all keys with a given prefix
func (rt *RadixTree) List(prefix string) ([]string, error) {
	result := []string{}

	// Handle empty prefix case - will return all keys
	if len(prefix) == 0 {
		err := rt.collectAllKeys(rt.RootID, "", &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	// Start from the root and find all matching keys
	err := rt.findKeysWithPrefix(rt.RootID, "", prefix, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Helper function to find all keys with a given prefix
func (rt *RadixTree) findKeysWithPrefix(nodeID uint32, currentPath, prefix string, result *[]string) error {
	nodeData, err := rt.DB.Get(nodeID)
	if err != nil {
		return err
	}
	node, err := deserializeNode(nodeData)
	if err != nil {
		return err
	}

	// If the current path already matches or exceeds the prefix length
	if len(currentPath) >= len(prefix) {
		// Check if the current path starts with the prefix
		if hasPrefix(currentPath, prefix) {
			// If this is a leaf node, add it to the results
			if node.IsLeaf {
				*result = append(*result, currentPath)
			}

			// Collect all keys from this subtree
			for _, child := range node.Children {
				childPath := currentPath + child.KeyPart
				err := rt.findKeysWithPrefix(child.NodeID, childPath, prefix, result)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Current path is shorter than the prefix, continue searching
	for _, child := range node.Children {
		childPath := currentPath + child.KeyPart

		// Check if this child's path could potentially match the prefix
		if hasPrefix(prefix, currentPath) {
			// The prefix starts with the current path, so we need to check if
			// the child's key_part matches the next part of the prefix
			prefixRemainder := prefix[len(currentPath):]

			// If the prefix remainder starts with the child's key_part or vice versa
			if hasPrefix(prefixRemainder, child.KeyPart) ||
				(hasPrefix(child.KeyPart, prefixRemainder) && len(child.KeyPart) >= len(prefixRemainder)) {
				err := rt.findKeysWithPrefix(child.NodeID, childPath, prefix, result)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Helper function to recursively collect all keys under a node
func (rt *RadixTree) collectAllKeys(nodeID uint32, currentPath string, result *[]string) error {
	nodeData, err := rt.DB.Get(nodeID)
	if err != nil {
		return err
	}
	node, err := deserializeNode(nodeData)
	if err != nil {
		return err
	}

	// If this node is a leaf, add its path to the result
	if node.IsLeaf {
		*result = append(*result, currentPath)
	}

	// Recursively collect keys from all children
	for _, child := range node.Children {
		childPath := currentPath + child.KeyPart
		err := rt.collectAllKeys(child.NodeID, childPath, result)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAll gets all values for keys with a given prefix
func (rt *RadixTree) GetAll(prefix string) ([][]byte, error) {
	// Get all matching keys
	keys, err := rt.List(prefix)
	if err != nil {
		return nil, err
	}

	// Get values for each key
	values := [][]byte{}
	for _, key := range keys {
		value, err := rt.Get(key)
		if err == nil {
			values = append(values, value)
		}
	}

	return values, nil
}

// Close closes the database
func (rt *RadixTree) Close() error {
	return rt.DB.Close()
}

// Destroy closes and removes the database
func (rt *RadixTree) Destroy() error {
	return rt.DB.Destroy()
}

// Helper function to get the common prefix of two strings
func getCommonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// Helper function to check if a string has a prefix
func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}
