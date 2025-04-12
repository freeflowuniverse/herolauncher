package radixtree

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const version = byte(1) // Current binary format version

// serializeNode serializes a node to bytes for storage
func serializeNode(node Node) []byte {
	// Calculate buffer size
	size := 1 + // version byte
		2 + len(node.KeySegment) + // key segment length (uint16) + data
		2 + len(node.Value) + // value length (uint16) + data
		2 // children count (uint16)

	// Add size for each child
	for _, child := range node.Children {
		size += 2 + len(child.KeyPart) + // key part length (uint16) + data
			4 // node ID (uint32)
	}

	size += 1 // leaf flag (byte)

	// Create buffer
	buf := make([]byte, 0, size)
	w := bytes.NewBuffer(buf)

	// Add version byte
	w.WriteByte(version)

	// Add key segment
	keySegmentLen := uint16(len(node.KeySegment))
	binary.Write(w, binary.LittleEndian, keySegmentLen)
	w.Write([]byte(node.KeySegment))

	// Add value
	valueLen := uint16(len(node.Value))
	binary.Write(w, binary.LittleEndian, valueLen)
	w.Write(node.Value)

	// Add children
	childrenLen := uint16(len(node.Children))
	binary.Write(w, binary.LittleEndian, childrenLen)
	for _, child := range node.Children {
		keyPartLen := uint16(len(child.KeyPart))
		binary.Write(w, binary.LittleEndian, keyPartLen)
		w.Write([]byte(child.KeyPart))
		binary.Write(w, binary.LittleEndian, child.NodeID)
	}

	// Add leaf flag
	if node.IsLeaf {
		w.WriteByte(1)
	} else {
		w.WriteByte(0)
	}

	return w.Bytes()
}

// deserializeNode deserializes bytes to a node
func deserializeNode(data []byte) (Node, error) {
	if len(data) < 1 {
		return Node{}, errors.New("data too short")
	}

	r := bytes.NewReader(data)

	// Read and verify version
	versionByte, err := r.ReadByte()
	if err != nil {
		return Node{}, err
	}
	if versionByte != version {
		return Node{}, errors.New("invalid version byte")
	}

	// Read key segment
	var keySegmentLen uint16
	if err := binary.Read(r, binary.LittleEndian, &keySegmentLen); err != nil {
		return Node{}, err
	}
	keySegmentBytes := make([]byte, keySegmentLen)
	if _, err := r.Read(keySegmentBytes); err != nil {
		return Node{}, err
	}
	keySegment := string(keySegmentBytes)

	// Read value
	var valueLen uint16
	if err := binary.Read(r, binary.LittleEndian, &valueLen); err != nil {
		return Node{}, err
	}
	value := make([]byte, valueLen)
	if _, err := r.Read(value); err != nil {
		return Node{}, err
	}

	// Read children
	var childrenLen uint16
	if err := binary.Read(r, binary.LittleEndian, &childrenLen); err != nil {
		return Node{}, err
	}
	children := make([]NodeRef, 0, childrenLen)
	for i := uint16(0); i < childrenLen; i++ {
		var keyPartLen uint16
		if err := binary.Read(r, binary.LittleEndian, &keyPartLen); err != nil {
			return Node{}, err
		}
		keyPartBytes := make([]byte, keyPartLen)
		if _, err := r.Read(keyPartBytes); err != nil {
			return Node{}, err
		}
		keyPart := string(keyPartBytes)

		var nodeID uint32
		if err := binary.Read(r, binary.LittleEndian, &nodeID); err != nil {
			return Node{}, err
		}

		children = append(children, NodeRef{
			KeyPart: keyPart,
			NodeID:  nodeID,
		})
	}

	// Read leaf flag
	isLeafByte, err := r.ReadByte()
	if err != nil {
		return Node{}, err
	}
	isLeaf := isLeafByte == 1

	return Node{
		KeySegment: keySegment,
		Value:      value,
		Children:   children,
		IsLeaf:     isLeaf,
	}, nil
}
