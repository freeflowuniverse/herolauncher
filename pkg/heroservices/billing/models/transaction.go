package models

import (
	"encoding/binary"
	"errors"
	"math"
)

// Transaction represents a financial transaction in the billing system
type Transaction struct {
	TxID    uint32  // Transaction ID, auto-incremented
	From    uint32  // Source account ID
	To      uint32  // Destination account ID
	Comment string  // Transaction comment
	Amount  float64 // Transaction amount
}

// Serialize converts a Transaction to a binary representation
func (t *Transaction) Serialize() []byte {
	// Calculate the size of the serialized data
	// 4 bytes for TxID + 4 bytes for From + 4 bytes for To +
	// 2 bytes for comment length + len(comment) bytes + 8 bytes for amount
	size := 4 + 4 + 4 + 2 + len(t.Comment) + 8
	data := make([]byte, size)

	// Write TxID (4 bytes)
	binary.LittleEndian.PutUint32(data[0:4], t.TxID)

	// Write From (4 bytes)
	binary.LittleEndian.PutUint32(data[4:8], t.From)

	// Write To (4 bytes)
	binary.LittleEndian.PutUint32(data[8:12], t.To)

	// Write comment length (2 bytes) and comment
	commentLen := uint16(len(t.Comment))
	binary.LittleEndian.PutUint16(data[12:14], commentLen)
	copy(data[14:14+commentLen], t.Comment)

	// Write amount (8 bytes)
	amountOffset := 14 + int(commentLen)
	binary.LittleEndian.PutUint64(data[amountOffset:amountOffset+8], math.Float64bits(t.Amount))

	return data
}

// DeserializeTransaction converts a binary representation back to a Transaction
func DeserializeTransaction(data []byte) (*Transaction, error) {
	if len(data) < 22 { // Minimum size: 4 (TxID) + 4 (From) + 4 (To) + 2 (comment length) + 8 (amount)
		return nil, errors.New("data too short to deserialize Transaction")
	}

	transaction := &Transaction{}

	// Read TxID
	transaction.TxID = binary.LittleEndian.Uint32(data[0:4])

	// Read From
	transaction.From = binary.LittleEndian.Uint32(data[4:8])

	// Read To
	transaction.To = binary.LittleEndian.Uint32(data[8:12])

	// Read comment length and comment
	commentLen := binary.LittleEndian.Uint16(data[12:14])
	if 14+commentLen > uint16(len(data)) {
		return nil, errors.New("data too short to read comment")
	}
	transaction.Comment = string(data[14 : 14+commentLen])

	// Read amount
	amountOffset := 14 + commentLen
	if int(amountOffset)+8 > len(data) {
		return nil, errors.New("data too short to read amount")
	}
	transaction.Amount = math.Float64frombits(binary.LittleEndian.Uint64(data[amountOffset : amountOffset+8]))

	return transaction, nil
}

// Validate checks if the transaction data is valid
func (t *Transaction) Validate() error {
	if t.From == 0 {
		return errors.New("source account ID cannot be zero")
	}

	if t.To == 0 {
		return errors.New("destination account ID cannot be zero")
	}

	if t.From == t.To {
		return errors.New("source and destination accounts cannot be the same")
	}

	if t.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	return nil
}
