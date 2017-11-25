package backdat

type ChunkStore interface {
	// Store the slice of ciphertext bytes by the provided key
	Store(string, []byte) error
	// Retrieve the slice of ciphertext bytes by the provided key
	Read(string) ([]byte, error)
	// Check if the ciphertext bytes are stored by the provided key
	Has(string) (bool, error)
	// List all keys
	List(chan string, chan error, chan interface{})
	// Delete a chunk
	Delete(string) error
}
