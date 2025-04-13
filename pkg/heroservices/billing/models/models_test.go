package models

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestEnvironment creates a temporary test environment
func setupTestEnvironment(t *testing.T) (string, func()) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "billing-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create subdirectories
	dbPath := filepath.Join(tempDir, "db")
	metaPath := filepath.Join(tempDir, "meta")

	if err := os.MkdirAll(dbPath, 0755); err != nil {
		t.Fatalf("Failed to create db directory: %v", err)
	}

	if err := os.MkdirAll(metaPath, 0755); err != nil {
		t.Fatalf("Failed to create meta directory: %v", err)
	}

	// Set environment variables to use the temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	// Return cleanup function
	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestUserCRUD tests the CRUD operations for User
func TestUserCRUD(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a new database store
	dbStore, err := NewDBStore()
	if err != nil {
		t.Fatalf("Failed to create database store: %v", err)
	}
	defer dbStore.Close()

	// Create a new user store
	userStore := &UserStore{Store: dbStore}

	// Create a new user
	user := &User{
		Key:      "john_doe",
		Accounts: []uint32{},
	}

	// Save the user
	err = userStore.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	// Verify the user ID was set
	if user.ID == 0 {
		t.Fatalf("User ID was not set")
	}

	// Get the user by ID
	retrievedUser, err := userStore.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	// Verify the user data
	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}
	if retrievedUser.Key != user.Key {
		t.Errorf("Expected user key %s, got %s", user.Key, retrievedUser.Key)
	}

	// Get the user by key
	retrievedUser, err = userStore.GetByKey("john_doe")
	if err != nil {
		t.Fatalf("Failed to get user by key: %v", err)
	}

	// Verify the user data
	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}

	// Update the user
	user.Key = "jane_doe"
	err = userStore.Save(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify the update
	retrievedUser, err = userStore.GetByKey("jane_doe")
	if err != nil {
		t.Fatalf("Failed to get updated user by key: %v", err)
	}
	if retrievedUser.Key != "jane_doe" {
		t.Errorf("Expected updated user key 'jane_doe', got %s", retrievedUser.Key)
	}

	// Delete the user
	err = userStore.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify the user was deleted
	_, err = userStore.GetByID(user.ID)
	if err == nil {
		t.Errorf("Expected error when getting deleted user, got nil")
	}
}

// TestAccountCRUD tests the CRUD operations for Account
func TestAccountCRUD(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a new database store
	dbStore, err := NewDBStore()
	if err != nil {
		t.Fatalf("Failed to create database store: %v", err)
	}
	defer dbStore.Close()

	// Create a new user store
	userStore := &UserStore{Store: dbStore}

	// Create a new user
	user := &User{
		Key:      "john_doe",
		Accounts: []uint32{},
	}

	// Save the user
	err = userStore.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	// Create a new account store
	accountStore := NewAccountStore(dbStore)

	// Create a new account
	account := &Account{
		UserID:   user.ID,
		Name:     "Savings",
		Currency: "USD",
		Amount:   1000.0,
	}

	// Save the account
	err = accountStore.Save(account)
	if err != nil {
		t.Fatalf("Failed to save account: %v", err)
	}

	// Verify the account ID was set
	if account.ID == 0 {
		t.Fatalf("Account ID was not set")
	}

	// Get the account by ID
	retrievedAccount, err := accountStore.GetByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get account by ID: %v", err)
	}

	// Verify the account data
	if retrievedAccount.ID != account.ID {
		t.Errorf("Expected account ID %d, got %d", account.ID, retrievedAccount.ID)
	}
	if retrievedAccount.UserID != user.ID {
		t.Errorf("Expected account user ID %d, got %d", user.ID, retrievedAccount.UserID)
	}
	if retrievedAccount.Name != "Savings" {
		t.Errorf("Expected account name 'Savings', got %s", retrievedAccount.Name)
	}
	if retrievedAccount.Currency != "USD" {
		t.Errorf("Expected account currency 'USD', got %s", retrievedAccount.Currency)
	}
	if retrievedAccount.Amount != 1000.0 {
		t.Errorf("Expected account amount 1000.0, got %f", retrievedAccount.Amount)
	}

	// Verify the user's accounts list was updated
	retrievedUser, err := userStore.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if len(retrievedUser.Accounts) != 1 {
		t.Errorf("Expected user to have 1 account, got %d", len(retrievedUser.Accounts))
	}
	if retrievedUser.Accounts[0] != account.ID {
		t.Errorf("Expected user account ID %d, got %d", account.ID, retrievedUser.Accounts[0])
	}

	// Update the account
	account.Name = "Checking"
	account.Amount = 2000.0
	err = accountStore.Save(account)
	if err != nil {
		t.Fatalf("Failed to update account: %v", err)
	}

	// Verify the update
	retrievedAccount, err = accountStore.GetByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get updated account: %v", err)
	}
	if retrievedAccount.Name != "Checking" {
		t.Errorf("Expected updated account name 'Checking', got %s", retrievedAccount.Name)
	}
	if retrievedAccount.Amount != 2000.0 {
		t.Errorf("Expected updated account amount 2000.0, got %f", retrievedAccount.Amount)
	}

	// Delete the account
	err = accountStore.Delete(account.ID)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	// Verify the account was deleted
	_, err = accountStore.GetByID(account.ID)
	if err == nil {
		t.Errorf("Expected error when getting deleted account, got nil")
	}

	// Verify the user's accounts list was updated
	retrievedUser, err = userStore.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user after account deletion: %v", err)
	}
	if len(retrievedUser.Accounts) != 0 {
		t.Errorf("Expected user to have 0 accounts after deletion, got %d", len(retrievedUser.Accounts))
	}
}

// TestTransactionCRUD tests the CRUD operations for Transaction
func TestTransactionCRUD(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a new database store
	dbStore, err := NewDBStore()
	if err != nil {
		t.Fatalf("Failed to create database store: %v", err)
	}
	defer dbStore.Close()

	// Create a new user store
	userStore := &UserStore{Store: dbStore}

	// Create a new user
	user := &User{
		Key:      "john_doe",
		Accounts: []uint32{},
	}

	// Save the user
	err = userStore.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	// Create a new account store
	accountStore := NewAccountStore(dbStore)

	// Create source account
	sourceAccount := &Account{
		UserID:   user.ID,
		Name:     "Checking",
		Currency: "USD",
		Amount:   1000.0,
	}

	// Save the source account
	err = accountStore.Save(sourceAccount)
	if err != nil {
		t.Fatalf("Failed to save source account: %v", err)
	}

	// Create destination account
	destAccount := &Account{
		UserID:   user.ID,
		Name:     "Savings",
		Currency: "USD",
		Amount:   500.0,
	}

	// Save the destination account
	err = accountStore.Save(destAccount)
	if err != nil {
		t.Fatalf("Failed to save destination account: %v", err)
	}

	// Create a new transaction store
	transactionStore := NewTransactionStore(dbStore)

	// Create a new transaction
	transaction := &Transaction{
		From:    sourceAccount.ID,
		To:      destAccount.ID,
		Comment: "Transfer to savings",
		Amount:  200.0,
	}

	// Save the transaction
	err = transactionStore.Save(transaction)
	if err != nil {
		t.Fatalf("Failed to save transaction: %v", err)
	}

	// Verify the transaction ID was set
	if transaction.TxID == 0 {
		t.Fatalf("Transaction ID was not set")
	}

	// Get the transaction by ID
	retrievedTransaction, err := transactionStore.GetByID(transaction.TxID)
	if err != nil {
		t.Fatalf("Failed to get transaction by ID: %v", err)
	}

	// Verify the transaction data
	if retrievedTransaction.TxID != transaction.TxID {
		t.Errorf("Expected transaction ID %d, got %d", transaction.TxID, retrievedTransaction.TxID)
	}
	if retrievedTransaction.From != sourceAccount.ID {
		t.Errorf("Expected transaction from %d, got %d", sourceAccount.ID, retrievedTransaction.From)
	}
	if retrievedTransaction.To != destAccount.ID {
		t.Errorf("Expected transaction to %d, got %d", destAccount.ID, retrievedTransaction.To)
	}
	if retrievedTransaction.Comment != "Transfer to savings" {
		t.Errorf("Expected transaction comment 'Transfer to savings', got %s", retrievedTransaction.Comment)
	}
	if retrievedTransaction.Amount != 200.0 {
		t.Errorf("Expected transaction amount 200.0, got %f", retrievedTransaction.Amount)
	}

	// Verify account balances were updated
	updatedSourceAccount, err := accountStore.GetByID(sourceAccount.ID)
	if err != nil {
		t.Fatalf("Failed to get updated source account: %v", err)
	}
	if updatedSourceAccount.Amount != 800.0 {
		t.Errorf("Expected source account amount 800.0, got %f", updatedSourceAccount.Amount)
	}

	updatedDestAccount, err := accountStore.GetByID(destAccount.ID)
	if err != nil {
		t.Fatalf("Failed to get updated destination account: %v", err)
	}
	if updatedDestAccount.Amount != 700.0 {
		t.Errorf("Expected destination account amount 700.0, got %f", updatedDestAccount.Amount)
	}

	// Verify getting transactions by account
	sourceTransactions, err := transactionStore.GetByAccount(sourceAccount.ID)
	if err != nil {
		t.Fatalf("Failed to get transactions by source account: %v", err)
	}
	if len(sourceTransactions) != 1 {
		t.Errorf("Expected 1 transaction for source account, got %d", len(sourceTransactions))
	}

	destTransactions, err := transactionStore.GetByAccount(destAccount.ID)
	if err != nil {
		t.Fatalf("Failed to get transactions by destination account: %v", err)
	}
	if len(destTransactions) != 1 {
		t.Errorf("Expected 1 transaction for destination account, got %d", len(destTransactions))
	}

	// Verify that deleting transactions is not allowed
	err = transactionStore.Delete(transaction.TxID)
	if err == nil {
		t.Errorf("Expected error when deleting transaction, got nil")
	}
}

// TestSerializationDeserialization tests the serialization and deserialization of models
func TestSerializationDeserialization(t *testing.T) {
	// Test User serialization/deserialization
	user := &User{
		ID:       42,
		Key:      "test_user",
		Accounts: []uint32{1, 2, 3},
	}

	userData := user.Serialize()
	deserializedUser, err := DeserializeUser(userData)
	if err != nil {
		t.Fatalf("Failed to deserialize user: %v", err)
	}

	if deserializedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, deserializedUser.ID)
	}
	if deserializedUser.Key != user.Key {
		t.Errorf("Expected user key %s, got %s", user.Key, deserializedUser.Key)
	}
	if len(deserializedUser.Accounts) != len(user.Accounts) {
		t.Errorf("Expected %d accounts, got %d", len(user.Accounts), len(deserializedUser.Accounts))
	}
	for i, accountID := range user.Accounts {
		if deserializedUser.Accounts[i] != accountID {
			t.Errorf("Expected account ID %d at index %d, got %d", accountID, i, deserializedUser.Accounts[i])
		}
	}

	// Test Account serialization/deserialization
	account := &Account{
		ID:       24,
		UserID:   42,
		Name:     "Test Account",
		Currency: "USD",
		Amount:   123.45,
	}

	accountData := account.Serialize()
	deserializedAccount, err := DeserializeAccount(accountData)
	if err != nil {
		t.Fatalf("Failed to deserialize account: %v", err)
	}

	if deserializedAccount.ID != account.ID {
		t.Errorf("Expected account ID %d, got %d", account.ID, deserializedAccount.ID)
	}
	if deserializedAccount.UserID != account.UserID {
		t.Errorf("Expected account user ID %d, got %d", account.UserID, deserializedAccount.UserID)
	}
	if deserializedAccount.Name != account.Name {
		t.Errorf("Expected account name %s, got %s", account.Name, deserializedAccount.Name)
	}
	if deserializedAccount.Currency != account.Currency {
		t.Errorf("Expected account currency %s, got %s", account.Currency, deserializedAccount.Currency)
	}
	if deserializedAccount.Amount != account.Amount {
		t.Errorf("Expected account amount %f, got %f", account.Amount, deserializedAccount.Amount)
	}

	// Test Transaction serialization/deserialization
	transaction := &Transaction{
		TxID:    123,
		From:    42,
		To:      24,
		Comment: "Test Transaction",
		Amount:  67.89,
	}

	transactionData := transaction.Serialize()
	deserializedTransaction, err := DeserializeTransaction(transactionData)
	if err != nil {
		t.Fatalf("Failed to deserialize transaction: %v", err)
	}

	if deserializedTransaction.TxID != transaction.TxID {
		t.Errorf("Expected transaction ID %d, got %d", transaction.TxID, deserializedTransaction.TxID)
	}
	if deserializedTransaction.From != transaction.From {
		t.Errorf("Expected transaction from %d, got %d", transaction.From, deserializedTransaction.From)
	}
	if deserializedTransaction.To != transaction.To {
		t.Errorf("Expected transaction to %d, got %d", transaction.To, deserializedTransaction.To)
	}
	if deserializedTransaction.Comment != transaction.Comment {
		t.Errorf("Expected transaction comment %s, got %s", transaction.Comment, deserializedTransaction.Comment)
	}
	if deserializedTransaction.Amount != transaction.Amount {
		t.Errorf("Expected transaction amount %f, got %f", transaction.Amount, deserializedTransaction.Amount)
	}
}
