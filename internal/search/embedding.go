// Package search provides hybrid search combining FTS5 keyword search with
// TF-IDF vector similarity for the observation memory system.
package search

import (
	"encoding/binary"
	"math"
	"strings"
	"unicode"
)

// stopWords are common English words filtered during tokenization.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true,
	"did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "shall": true, "can": true,
	"it": true, "its": true, "this": true, "that": true, "these": true,
	"those": true, "i": true, "you": true, "he": true, "she": true,
	"we": true, "they": true, "me": true, "him": true, "her": true,
	"us": true, "them": true, "my": true, "your": true, "his": true,
	"our": true, "their": true,
	"in": true, "on": true, "at": true, "to": true, "for": true,
	"of": true, "with": true, "by": true, "from": true, "as": true,
	"into": true, "about": true, "between": true,
	"and": true, "or": true, "not": true, "but": true, "if": true,
	"so": true, "no": true, "nor": true,
}

// tokenize splits text into lowercase tokens, removing stop words and
// non-alphabetic characters.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	var tokens []string
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if stopWords[w] {
			continue
		}
		tokens = append(tokens, w)
	}
	return tokens
}

// Vocabulary maps tokens to vector indices and stores IDF weights.
type Vocabulary struct {
	index map[string]int // token → vector index
	idf   []float64      // IDF weight per index
	size  int
}

// NewVocabulary builds a vocabulary from a corpus of documents.
// Each document is a raw text string. Computes IDF weights.
func NewVocabulary(docs []string) *Vocabulary {
	// Count document frequency for each token
	df := make(map[string]int)
	for _, doc := range docs {
		seen := make(map[string]bool)
		for _, tok := range tokenize(doc) {
			if !seen[tok] {
				df[tok]++
				seen[tok] = true
			}
		}
	}

	// Build index and compute IDF
	v := &Vocabulary{
		index: make(map[string]int, len(df)),
		idf:   make([]float64, len(df)),
		size:  len(df),
	}

	n := float64(len(docs))
	i := 0
	for tok, count := range df {
		v.index[tok] = i
		// IDF = log(N / df) + 1 (smoothed)
		v.idf[i] = math.Log(n/float64(count)) + 1.0
		i++
	}

	return v
}

// Size returns the dimensionality of the embedding vectors.
func (v *Vocabulary) Size() int {
	return v.size
}

// IDF returns the IDF weight for a token. Returns 0 for unknown tokens.
func (v *Vocabulary) IDF(token string) float64 {
	token = strings.ToLower(token)
	if idx, ok := v.index[token]; ok {
		return v.idf[idx]
	}
	return 0
}

// Embed converts text into a TF-IDF vector using this vocabulary.
// The vector is L2-normalized.
func (v *Vocabulary) Embed(text string) []float64 {
	vec := make([]float64, v.size)
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return vec
	}

	// Count term frequency
	tf := make(map[string]int)
	for _, tok := range tokens {
		tf[tok]++
	}

	// Compute TF-IDF
	for tok, count := range tf {
		if idx, ok := v.index[tok]; ok {
			// TF = count / total tokens (normalized)
			tfNorm := float64(count) / float64(len(tokens))
			vec[idx] = tfNorm * v.idf[idx]
		}
	}

	// L2 normalize
	normalize(vec)
	return vec
}

// normalize applies L2 normalization in-place.
func normalize(vec []float64) {
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return
	}
	for i := range vec {
		vec[i] /= norm
	}
}

// CosineSimilarity computes the cosine similarity between two vectors.
// Returns 0 for empty or zero-length vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// EncodeVector serializes a float64 slice to bytes for SQLite BLOB storage.
func EncodeVector(vec []float64) []byte {
	buf := make([]byte, len(vec)*8)
	for i, v := range vec {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
	}
	return buf
}

// DecodeVector deserializes bytes from SQLite BLOB back to a float64 slice.
func DecodeVector(data []byte) []float64 {
	if len(data) == 0 {
		return nil
	}
	n := len(data) / 8
	vec := make([]float64, n)
	for i := 0; i < n; i++ {
		vec[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[i*8:]))
	}
	return vec
}
