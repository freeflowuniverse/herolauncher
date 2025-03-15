package ourdb

import (
	"errors"
)

// Client provides a simplified interface to the OurDB database
type Client struct {
	db *OurDB
}

// NewClient creates a new client for the specified database path
func NewClient(path string) (*Client, error) {
	return NewClientWithConfig(path, DefaultConfig())
}

// NewClientWithConfig creates a new client with a custom configuration
func NewClientWithConfig(path string, baseConfig OurDBConfig) (*Client, error) {
	config := baseConfig
	config.Path = path

	db, err := New(config)
	if err != nil {
		return nil, err
	}

	return &Client{db: db}, nil
}

// Set stores data with the specified ID
func (c *Client) Set(id uint32, data []byte) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}

	_, err := c.db.Set(OurDBSetArgs{
		ID:   &id,
		Data: data,
	})
	return err
}

// Add stores data and returns the auto-generated ID
func (c *Client) Add(data []byte) (uint32, error) {
	if data == nil {
		return 0, errors.New("data cannot be nil")
	}

	return c.db.Set(OurDBSetArgs{
		Data: data,
	})
}

// Get retrieves data for the specified ID
func (c *Client) Get(id uint32) ([]byte, error) {
	return c.db.Get(id)
}

// GetHistory retrieves historical values for the specified ID
func (c *Client) GetHistory(id uint32, depth uint8) ([][]byte, error) {
	return c.db.GetHistory(id, depth)
}

// Delete removes data for the specified ID
func (c *Client) Delete(id uint32) error {
	return c.db.Delete(id)
}

// Close closes the database
func (c *Client) Close() error {
	return c.db.Close()
}

// Destroy closes and removes the database
func (c *Client) Destroy() error {
	return c.db.Destroy()
}
