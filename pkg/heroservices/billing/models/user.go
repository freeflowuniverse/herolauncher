package models

import (
	"encoding/binary"
	"errors"
)

// User represents a user in the billing system
type User struct {
	ID       uint32   // Unique identifier, auto-incremented
	Key      string   // Unique key for the user
	Accounts []uint32 // Links to the accounts a user has
}

// Serialize converts a User to a binary representation
func (u *User) Serialize() []byte {
	// Calculate the size of the serialized data
	// 4 bytes for ID + 2 bytes for key length + len(key) bytes + 2 bytes for accounts length + 4 bytes per account
	size := 4 + 2 + len(u.Key) + 2 + (4 * len(u.Accounts))
	data := make([]byte, size)

	// Write ID (4 bytes)
	binary.LittleEndian.PutUint32(data[0:4], u.ID)

	// Write key length (2 bytes) and key
	keyLen := uint16(len(u.Key))
	binary.LittleEndian.PutUint16(data[4:6], keyLen)
	copy(data[6:6+keyLen], u.Key)

	// Write accounts length (2 bytes) and accounts
	accountsOffset := 6 + keyLen
	accountsLen := uint16(len(u.Accounts))
	binary.LittleEndian.PutUint16(data[accountsOffset:accountsOffset+2], accountsLen)

	for i, accountID := range u.Accounts {
		offset := int(accountsOffset) + 2 + (4 * i)
		binary.LittleEndian.PutUint32(data[offset:offset+4], accountID)
	}

	return data
}

// DeserializeUser converts a binary representation back to a User
func DeserializeUser(data []byte) (*User, error) {
	if len(data) < 8 { // Minimum size: 4 (ID) + 2 (key length) + 2 (accounts length)
		return nil, errors.New("data too short to deserialize User")
	}

	user := &User{}

	// Read ID
	user.ID = binary.LittleEndian.Uint32(data[0:4])

	// Read key length and key
	keyLen := binary.LittleEndian.Uint16(data[4:6])
	if 6+keyLen > uint16(len(data)) {
		return nil, errors.New("data too short to read key")
	}
	user.Key = string(data[6 : 6+keyLen])

	// Read accounts length and accounts
	accountsOffset := 6 + keyLen
	if accountsOffset+2 > uint16(len(data)) {
		return nil, errors.New("data too short to read accounts length")
	}
	accountsLen := binary.LittleEndian.Uint16(data[accountsOffset : accountsOffset+2])

	user.Accounts = make([]uint32, accountsLen)
	for i := uint16(0); i < accountsLen; i++ {
		offset := accountsOffset + 2 + (4 * i)
		if offset+4 > uint16(len(data)) {
			return nil, errors.New("data too short to read account ID")
		}
		user.Accounts[i] = binary.LittleEndian.Uint32(data[offset : offset+4])
	}

	return user, nil
}
