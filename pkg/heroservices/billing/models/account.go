package models

import (
	"encoding/binary"
	"errors"
	"math"
	"strings"
)

// Account represents a financial account in the billing system
type Account struct {
	ID       uint32  // Unique identifier, auto-incremented
	UserID   uint32  // Link to the unique user ID
	Name     string  // Name of the account
	Currency string  // Currency code (e.g., USD, EUR), always uppercase, 3 letters
	Amount   float64 // Current balance, always positive
}

// Serialize converts an Account to a binary representation
func (a *Account) Serialize() []byte {
	// Calculate the size of the serialized data
	// 4 bytes for ID + 4 bytes for UserID + 2 bytes for name length + len(name) bytes +
	// 3 bytes for currency + 8 bytes for amount
	size := 4 + 4 + 2 + len(a.Name) + 3 + 8
	data := make([]byte, size)

	// Write ID (4 bytes)
	binary.LittleEndian.PutUint32(data[0:4], a.ID)

	// Write UserID (4 bytes)
	binary.LittleEndian.PutUint32(data[4:8], a.UserID)

	// Write name length (2 bytes) and name
	nameLen := uint16(len(a.Name))
	binary.LittleEndian.PutUint16(data[8:10], nameLen)
	copy(data[10:10+nameLen], a.Name)

	// Write currency (3 bytes)
	currencyOffset := 10 + nameLen
	copy(data[currencyOffset:currencyOffset+3], a.Currency)

	// Write amount (8 bytes)
	amountOffset := currencyOffset + 3
	binary.LittleEndian.PutUint64(data[amountOffset:amountOffset+8], math.Float64bits(a.Amount))

	return data
}

// DeserializeAccount converts a binary representation back to an Account
func DeserializeAccount(data []byte) (*Account, error) {
	if len(data) < 17 { // Minimum size: 4 (ID) + 4 (UserID) + 2 (name length) + 3 (currency) + 8 (amount)
		return nil, errors.New("data too short to deserialize Account")
	}

	account := &Account{}

	// Read ID
	account.ID = binary.LittleEndian.Uint32(data[0:4])

	// Read UserID
	account.UserID = binary.LittleEndian.Uint32(data[4:8])

	// Read name length and name
	nameLen := binary.LittleEndian.Uint16(data[8:10])
	if 10+nameLen > uint16(len(data)) {
		return nil, errors.New("data too short to read name")
	}
	account.Name = string(data[10 : 10+nameLen])

	// Read currency
	currencyOffset := 10 + nameLen
	if int(currencyOffset)+3 > len(data) {
		return nil, errors.New("data too short to read currency")
	}
	account.Currency = string(data[currencyOffset : currencyOffset+3])

	// Read amount
	amountOffset := currencyOffset + 3
	if int(amountOffset)+8 > len(data) {
		return nil, errors.New("data too short to read amount")
	}
	account.Amount = math.Float64frombits(binary.LittleEndian.Uint64(data[amountOffset : amountOffset+8]))

	return account, nil
}

// Validate checks if the account data is valid
func (a *Account) Validate() error {
	if a.UserID == 0 {
		return errors.New("user ID cannot be zero")
	}

	if a.Name == "" {
		return errors.New("account name cannot be empty")
	}

	// Check currency format (3 uppercase letters)
	if len(a.Currency) != 3 {
		return errors.New("currency must be 3 characters")
	}

	a.Currency = strings.ToUpper(a.Currency)
	for _, r := range a.Currency {
		if r < 'A' || r > 'Z' {
			return errors.New("currency must contain only uppercase letters")
		}
	}

	// Ensure amount is non-negative
	if a.Amount < 0 {
		return errors.New("amount cannot be negative")
	}

	return nil
}
