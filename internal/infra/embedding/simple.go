package embedding

import (
	"math"
	"strings"
)

const dims = 384

// SimpleEmbedder is a lightweight TF-IDF-style embedder that produces
// deterministic 384-dim float64 vectors. It is intended as a placeholder
// until a real embedding API is wired in.
type SimpleEmbedder struct{}

// NewSimpleEmbedder returns a SimpleEmbedder.
func NewSimpleEmbedder() *SimpleEmbedder { return &SimpleEmbedder{} }

// Dims returns the vector dimensionality.
func (e *SimpleEmbedder) Dims() int { return dims }

// Embed converts text to a fixed-length vector using character n-gram hashing.
func (e *SimpleEmbedder) Embed(text string) ([]float64, error) {
	vec := make([]float64, dims)
	text = strings.ToLower(text)
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		// Unigram
		h := hash(string(runes[i]))
		vec[h%dims] += 1.0

		// Bigram
		if i+1 < len(runes) {
			h2 := hash(string(runes[i:i+2]))
			vec[h2%dims] += 0.7
		}

		// Trigram
		if i+2 < len(runes) {
			h3 := hash(string(runes[i:i+3]))
			vec[h3%dims] += 0.5
		}
	}

	// L2-normalise
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}

	return vec, nil
}

// hash is a simple FNV-1a 32-bit hash returning a non-negative integer.
func hash(s string) int {
	const (
		prime  = 16777619
		offset = 2166136261
	)
	h := uint32(offset)
	for _, b := range []byte(s) {
		h ^= uint32(b)
		h *= prime
	}
	return int(h & 0x7fffffff)
}
