package cache

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"

	"github.com/redis/go-redis/v9"
)

// BloomFilter implements a Redis-backed bloom filter
type BloomFilter struct {
	client *redis.Client
	key    string
	size   uint64 // Number of bits
	hashes uint   // Number of hash functions
}

// NewBloomFilter creates a new bloom filter
// expectedItems: expected number of items to be added
// falsePositiveRate: desired false positive rate (e.g., 0.01 for 1%)
func NewBloomFilter(client *redis.Client, key string, expectedItems uint64, falsePositiveRate float64) *BloomFilter {
	// Calculate optimal size and number of hash functions
	size := optimalSize(expectedItems, falsePositiveRate)
	hashes := optimalHashes(size, expectedItems)

	return &BloomFilter{
		client: client,
		key:    key,
		size:   size,
		hashes: hashes,
	}
}

// Add adds an item to the bloom filter
func (bf *BloomFilter) Add(ctx context.Context, item string) error {
	positions := bf.getPositions(item)

	pipe := bf.client.Pipeline()
	for _, pos := range positions {
		pipe.SetBit(ctx, bf.key, int64(pos), 1)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// AddMultiple adds multiple items to the bloom filter
func (bf *BloomFilter) AddMultiple(ctx context.Context, items []string) error {
	pipe := bf.client.Pipeline()
	for _, item := range items {
		positions := bf.getPositions(item)
		for _, pos := range positions {
			pipe.SetBit(ctx, bf.key, int64(pos), 1)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Exists checks if an item might exist in the bloom filter
// Returns true if the item might exist (with false positive rate)
// Returns false if the item definitely does not exist
func (bf *BloomFilter) Exists(ctx context.Context, item string) (bool, error) {
	positions := bf.getPositions(item)

	pipe := bf.client.Pipeline()
	cmds := make([]*redis.IntCmd, len(positions))
	for i, pos := range positions {
		cmds[i] = pipe.GetBit(ctx, bf.key, int64(pos))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	for _, cmd := range cmds {
		if cmd.Val() == 0 {
			return false, nil
		}
	}

	return true, nil
}

// Clear removes all items from the bloom filter
func (bf *BloomFilter) Clear(ctx context.Context) error {
	return bf.client.Del(ctx, bf.key).Err()
}

// getPositions calculates the bit positions for an item
func (bf *BloomFilter) getPositions(item string) []uint64 {
	positions := make([]uint64, bf.hashes)

	// Use SHA256 to generate hash values
	hash := sha256.Sum256([]byte(item))

	// Split the hash into two 64-bit values for double hashing
	h1 := binary.BigEndian.Uint64(hash[:8])
	h2 := binary.BigEndian.Uint64(hash[8:16])

	// Generate k hash values using double hashing
	for i := uint(0); i < bf.hashes; i++ {
		positions[i] = (h1 + uint64(i)*h2) % bf.size
	}

	return positions
}

// optimalSize calculates the optimal bit array size
func optimalSize(n uint64, p float64) uint64 {
	// m = -(n * ln(p)) / (ln(2)^2)
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	return uint64(math.Ceil(m))
}

// optimalHashes calculates the optimal number of hash functions
func optimalHashes(m, n uint64) uint {
	// k = (m/n) * ln(2)
	k := float64(m) / float64(n) * math.Ln2
	return uint(math.Ceil(k))
}

// Info returns information about the bloom filter
func (bf *BloomFilter) Info() map[string]any {
	return map[string]any{
		"size":   bf.size,
		"hashes": bf.hashes,
		"key":    bf.key,
	}
}
