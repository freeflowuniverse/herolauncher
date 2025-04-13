package models

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/data/ourdb"
	"github.com/freeflowuniverse/herolauncher/pkg/data/radixtree"
	"github.com/freeflowuniverse/herolauncher/pkg/tools"
)

// DBStore represents the central database store for all models
type DBStore struct {
	DB   *ourdb.OurDB
	Meta *radixtree.RadixTree
}

// NewDBStore creates a new database store
func NewDBStore() (*DBStore, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create paths
	dbPath := filepath.Join(homeDir, "hero", "var", "server", "users", "db")
	metaPath := filepath.Join(homeDir, "hero", "var", "server", "users", "meta")

	// Create directories if they don't exist
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}
	if err := os.MkdirAll(metaPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create meta directory: %w", err)
	}

	// Create ourdb
	dbConfig := ourdb.DefaultConfig()
	dbConfig.Path = dbPath
	dbConfig.IncrementalMode = true
	db, err := ourdb.New(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ourdb: %w", err)
	}

	// Create radixtree
	meta, err := radixtree.New(radixtree.NewArgs{
		Path: metaPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create radixtree: %w", err)
	}

	return &DBStore{
		DB:   db,
		Meta: meta,
	}, nil
}

// Close closes the database connections
func (s *DBStore) Close() error {
	err1 := s.DB.Close()
	err2 := s.Meta.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

// UserStore handles the storage and retrieval of User objects
type UserStore struct {
	Store *DBStore
}

// NewUserStore creates a new UserStore
func NewUserStore() (*UserStore, *DBStore, error) {
	store, err := NewDBStore()
	if err != nil {
		return nil, nil, err
	}

	return &UserStore{
		Store: store,
	}, store, nil
}

// Save saves a User to the database
func (s *UserStore) Save(user *User) error {
	// If ID is 0, this is a new user
	if user.ID == 0 {
		// Fix the key
		fixedKey := tools.NameFix(user.Key)
		if fixedKey == "" {
			return errors.New("key cannot be empty")
		}
		user.Key = fixedKey

		// Check if key already exists
		existingID, err := s.GetIDByKey(user.Key)
		if err == nil && existingID != 0 {
			return fmt.Errorf("user with key %s already exists", user.Key)
		}

		// Get next ID
		nextID, err := s.Store.DB.GetNextID()
		if err != nil {
			return fmt.Errorf("failed to get next ID: %w", err)
		}
		user.ID = nextID

		// Save to ourdb
		data := user.Serialize()
		_, err = s.Store.DB.Set(ourdb.OurDBSetArgs{
			Data: data,
		})
		if err != nil {
			return fmt.Errorf("failed to save user to ourdb: %w", err)
		}

		// Save key to radixtree
		idBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(idBytes, user.ID)
		err = s.Store.Meta.Set(user.Key, idBytes)
		if err != nil {
			return fmt.Errorf("failed to save user key to radixtree: %w", err)
		}
	} else {
		// Update existing user
		// Get existing user to verify key
		existingUser, err := s.GetByID(user.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing user: %w", err)
		}

		// If key changed, update radixtree
		if existingUser.Key != user.Key {
			// Fix the new key
			fixedKey := tools.NameFix(user.Key)
			if fixedKey == "" {
				return errors.New("key cannot be empty")
			}
			user.Key = fixedKey

			// Check if new key already exists
			existingID, err := s.GetIDByKey(user.Key)
			if err == nil && existingID != 0 && existingID != user.ID {
				return fmt.Errorf("user with key %s already exists", user.Key)
			}

			// Delete old key from radixtree
			err = s.Store.Meta.Delete(existingUser.Key)
			if err != nil {
				return fmt.Errorf("failed to delete old user key from radixtree: %w", err)
			}

			// Save new key to radixtree
			idBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(idBytes, user.ID)
			err = s.Store.Meta.Set(user.Key, idBytes)
			if err != nil {
				return fmt.Errorf("failed to save new user key to radixtree: %w", err)
			}
		}

		// Save to ourdb
		data := user.Serialize()
		id := user.ID
		_, err = s.Store.DB.Set(ourdb.OurDBSetArgs{
			ID:   &id,
			Data: data,
		})
		if err != nil {
			return fmt.Errorf("failed to update user in ourdb: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a User by ID
func (s *UserStore) GetByID(id uint32) (*User, error) {
	data, err := s.Store.DB.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from ourdb: %w", err)
	}

	user, err := DeserializeUser(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize user: %w", err)
	}

	return user, nil
}

// GetByKey retrieves a User by key
func (s *UserStore) GetByKey(key string) (*User, error) {
	// Fix the key
	fixedKey := tools.NameFix(key)
	if fixedKey == "" {
		return nil, errors.New("key cannot be empty")
	}

	// Get ID from radixtree
	id, err := s.GetIDByKey(fixedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from radixtree: %w", err)
	}

	// Get user from ourdb
	return s.GetByID(id)
}

// GetIDByKey retrieves a user ID by key
func (s *UserStore) GetIDByKey(key string) (uint32, error) {
	// Get ID from radixtree
	idBytes, err := s.Store.Meta.Get(key)
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID from radixtree: %w", err)
	}

	if len(idBytes) != 4 {
		return 0, fmt.Errorf("invalid ID bytes length: %d", len(idBytes))
	}

	return binary.LittleEndian.Uint32(idBytes), nil
}

// Delete deletes a User by ID
func (s *UserStore) Delete(id uint32) error {
	// Get user to get key
	user, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Delete from radixtree
	err = s.Store.Meta.Delete(user.Key)
	if err != nil {
		return fmt.Errorf("failed to delete user key from radixtree: %w", err)
	}

	// Delete from ourdb
	err = s.Store.DB.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete user from ourdb: %w", err)
	}

	return nil
}

// AccountStore handles the storage and retrieval of Account objects
type AccountStore struct {
	Store *DBStore
}

// NewAccountStore creates a new AccountStore using the provided DBStore
func NewAccountStore(store *DBStore) *AccountStore {
	return &AccountStore{
		Store: store,
	}
}

// Save saves an Account to the database
func (s *AccountStore) Save(account *Account) error {
	// Validate account data
	if err := account.Validate(); err != nil {
		return err
	}

	// If ID is 0, this is a new account
	if account.ID == 0 {
		// Get next ID
		nextID, err := s.Store.DB.GetNextID()
		if err != nil {
			return fmt.Errorf("failed to get next ID: %w", err)
		}
		account.ID = nextID

		// Save to ourdb
		data := account.Serialize()
		_, err = s.Store.DB.Set(ourdb.OurDBSetArgs{
			Data: data,
		})
		if err != nil {
			return fmt.Errorf("failed to save account to ourdb: %w", err)
		}

		// Update user's accounts list
		userStore := &UserStore{Store: s.Store}
		user, err := userStore.GetByID(account.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Add account ID to user's accounts
		user.Accounts = append(user.Accounts, account.ID)

		// Save updated user
		err = userStore.Save(user)
		if err != nil {
			return fmt.Errorf("failed to update user's accounts: %w", err)
		}
	} else {
		// Update existing account
		// Get existing account to verify it exists
		_, err := s.GetByID(account.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing account: %w", err)
		}

		// Save to ourdb
		data := account.Serialize()
		id := account.ID
		_, err = s.Store.DB.Set(ourdb.OurDBSetArgs{
			ID:   &id,
			Data: data,
		})
		if err != nil {
			return fmt.Errorf("failed to update account in ourdb: %w", err)
		}
	}

	return nil
}

// GetByID retrieves an Account by ID
func (s *AccountStore) GetByID(id uint32) (*Account, error) {
	data, err := s.Store.DB.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account from ourdb: %w", err)
	}

	account, err := DeserializeAccount(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize account: %w", err)
	}

	return account, nil
}

// GetByUserID retrieves all Accounts for a user
func (s *AccountStore) GetByUserID(userID uint32) ([]*Account, error) {
	// Get user to get account IDs
	userStore := &UserStore{Store: s.Store}
	user, err := userStore.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get accounts
	accounts := make([]*Account, 0, len(user.Accounts))
	for _, accountID := range user.Accounts {
		account, err := s.GetByID(accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account %d: %w", accountID, err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Delete deletes an Account by ID
func (s *AccountStore) Delete(id uint32) error {
	// Get account to get user ID
	account, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Get user to remove account from accounts list
	userStore := &UserStore{Store: s.Store}
	user, err := userStore.GetByID(account.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Remove account ID from user's accounts
	for i, accountID := range user.Accounts {
		if accountID == id {
			user.Accounts = append(user.Accounts[:i], user.Accounts[i+1:]...)
			break
		}
	}

	// Save updated user
	err = userStore.Save(user)
	if err != nil {
		return fmt.Errorf("failed to update user's accounts: %w", err)
	}

	// Delete from ourdb
	err = s.Store.DB.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete account from ourdb: %w", err)
	}

	return nil
}

// TransactionStore handles the storage and retrieval of Transaction objects
type TransactionStore struct {
	Store *DBStore
}

// NewTransactionStore creates a new TransactionStore using the provided DBStore
func NewTransactionStore(store *DBStore) *TransactionStore {
	return &TransactionStore{
		Store: store,
	}
}

// Save saves a Transaction to the database and updates account balances
func (s *TransactionStore) Save(transaction *Transaction) error {
	// Validate transaction data
	if err := transaction.Validate(); err != nil {
		return err
	}

	// Get account store
	accountStore := &AccountStore{Store: s.Store}

	// Get source account
	fromAccount, err := accountStore.GetByID(transaction.From)
	if err != nil {
		return fmt.Errorf("failed to get source account: %w", err)
	}

	// Get destination account
	toAccount, err := accountStore.GetByID(transaction.To)
	if err != nil {
		return fmt.Errorf("failed to get destination account: %w", err)
	}

	// Check if source account has enough balance
	if fromAccount.Amount < transaction.Amount {
		return errors.New("insufficient balance in source account")
	}

	// If TxID is 0, this is a new transaction
	if transaction.TxID == 0 {
		// Get next ID
		nextID, err := s.Store.DB.GetNextID()
		if err != nil {
			return fmt.Errorf("failed to get next ID: %w", err)
		}
		transaction.TxID = nextID

		// Update account balances
		fromAccount.Amount -= transaction.Amount
		toAccount.Amount += transaction.Amount

		// Save updated accounts
		if err := accountStore.Save(fromAccount); err != nil {
			return fmt.Errorf("failed to update source account: %w", err)
		}

		if err := accountStore.Save(toAccount); err != nil {
			// Rollback source account change if destination update fails
			fromAccount.Amount += transaction.Amount
			if rollbackErr := accountStore.Save(fromAccount); rollbackErr != nil {
				return fmt.Errorf("failed to update destination account and rollback failed: %w, rollback error: %v", err, rollbackErr)
			}
			return fmt.Errorf("failed to update destination account: %w", err)
		}

		// Save transaction to ourdb
		data := transaction.Serialize()
		_, err = s.Store.DB.Set(ourdb.OurDBSetArgs{
			Data: data,
		})
		if err != nil {
			// Rollback account changes if transaction save fails
			fromAccount.Amount += transaction.Amount
			toAccount.Amount -= transaction.Amount
			if rollbackErr1 := accountStore.Save(fromAccount); rollbackErr1 != nil {
				return fmt.Errorf("failed to save transaction and rollback failed: %w, rollback error 1: %v", err, rollbackErr1)
			}
			if rollbackErr2 := accountStore.Save(toAccount); rollbackErr2 != nil {
				return fmt.Errorf("failed to save transaction and rollback failed: %w, rollback error 2: %v", err, rollbackErr2)
			}
			return fmt.Errorf("failed to save transaction to ourdb: %w", err)
		}
	} else {
		// Updating existing transactions is not allowed as it would affect account balances
		return errors.New("updating existing transactions is not allowed")
	}

	return nil
}

// GetByID retrieves a Transaction by ID
func (s *TransactionStore) GetByID(id uint32) (*Transaction, error) {
	data, err := s.Store.DB.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction from ourdb: %w", err)
	}

	transaction, err := DeserializeTransaction(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %w", err)
	}

	return transaction, nil
}

// GetByAccount retrieves all Transactions for an account (either as source or destination)
func (s *TransactionStore) GetByAccount(accountID uint32) ([]*Transaction, error) {
	// This is a simplified implementation that would be inefficient in a real system
	// In a real system, we would maintain indexes for account transactions

	// Get the next ID to determine the range of transaction IDs
	nextID, err := s.Store.DB.GetNextID()
	if err != nil {
		return nil, fmt.Errorf("failed to get next ID: %w", err)
	}

	transactions := make([]*Transaction, 0)

	// Iterate through all transactions (this is inefficient but works for a simple implementation)
	for i := uint32(1); i < nextID; i++ {
		transaction, err := s.GetByID(i)
		if err != nil {
			// Skip if transaction doesn't exist or can't be deserialized
			continue
		}

		// Include transaction if it involves the specified account
		if transaction.From == accountID || transaction.To == accountID {
			transactions = append(transactions, transaction)
		}
	}

	return transactions, nil
}

// Delete is not implemented for transactions as it would affect account balances
// In a real system, you might want to implement a "void" or "reverse" transaction instead
func (s *TransactionStore) Delete(id uint32) error {
	return errors.New("deleting transactions is not allowed")
}
