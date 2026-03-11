package embedding

// Embedder computes a vector embedding for a text string.
type Embedder interface {
	// Embed returns a fixed-length float64 slice representing the input text.
	Embed(text string) ([]float64, error)
	// Dims returns the dimensionality of the embedding vectors.
	Dims() int
}
