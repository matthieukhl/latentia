package embed

import (
	"context"
	"crypto/md5"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/matthieukhl/latentia/internal/types"
)

type MockEmbedder struct {
	model string
	dim   int
}

func NewMockEmbedder(model string, dim int) *MockEmbedder {
	return &MockEmbedder{
		model: model,
		dim:   dim,
	}
}

func (e *MockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		// Create deterministic "embeddings" based on text content
		embedding := e.generateDeterministicEmbedding(text)
		embeddings[i] = embedding
	}
	
	// Simulate API delay
	time.Sleep(100 * time.Millisecond)
	
	return embeddings, nil
}

func (e *MockEmbedder) Dim() int {
	return e.dim
}

func (e *MockEmbedder) Model() string {
	return e.model + "-mock"
}

// generateDeterministicEmbedding creates a deterministic embedding based on text content
func (e *MockEmbedder) generateDeterministicEmbedding(text string) []float32 {
	// Create a hash of the text for deterministic seeding
	hash := md5.Sum([]byte(text))
	seed := int64(0)
	for i := 0; i < 8; i++ {
		seed = seed<<8 + int64(hash[i])
	}
	
	rng := rand.New(rand.NewSource(seed))
	embedding := make([]float32, e.dim)
	
	// Generate embedding with some structure based on text content
	text = strings.ToLower(text)
	
	// Base random values
	for i := 0; i < e.dim; i++ {
		embedding[i] = rng.Float32()*2 - 1 // Random between -1 and 1
	}
	
	// Add semantic-like features for SQL terms
	sqlTerms := map[string]int{
		"select": 0, "from": 1, "where": 2, "join": 3, "index": 4,
		"order": 5, "group": 6, "having": 7, "limit": 8, "union": 9,
		"insert": 10, "update": 11, "delete": 12, "create": 13, "alter": 14,
		"optimize": 15, "performance": 16, "slow": 17, "fast": 18, "query": 19,
		"table": 20, "column": 21, "primary": 22, "foreign": 23, "key": 24,
		"aggregate": 25, "count": 26, "sum": 27, "avg": 28, "max": 29, "min": 30,
	}
	
	// Boost dimensions based on SQL keywords
	for term, dim := range sqlTerms {
		if strings.Contains(text, term) && dim < e.dim {
			embedding[dim] += 0.5
		}
	}
	
	// Normalize to unit vector
	var norm float32 = 0
	for _, val := range embedding {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))
	
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}
	
	return embedding
}

// Compile-time interface check
var _ types.Embedder = (*MockEmbedder)(nil)