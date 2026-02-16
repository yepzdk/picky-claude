package search

import (
	"math"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  int // expected token count (approximate)
	}{
		{"hello world", 2},
		{"Hello, World!", 2},
		{"the a is an", 0}, // all stop words
		{"authentication flow is broken", 3},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(tt.input)
			if len(tokens) != tt.want {
				t.Errorf("tokenize(%q) = %d tokens %v, want %d", tt.input, len(tokens), tokens, tt.want)
			}
		})
	}
}

func TestNewVocabulary(t *testing.T) {
	docs := []string{
		"authentication flow broken",
		"database migration failed",
		"authentication database connection",
	}

	vocab := NewVocabulary(docs)
	if vocab.Size() == 0 {
		t.Fatal("vocabulary is empty")
	}

	// "authentication" appears in 2 of 3 docs, should have lower IDF than "migration"
	authIDF := vocab.IDF("authentication")
	migIDF := vocab.IDF("migration")
	if authIDF >= migIDF {
		t.Errorf("IDF(authentication)=%f should be < IDF(migration)=%f", authIDF, migIDF)
	}
}

func TestEmbed(t *testing.T) {
	docs := []string{
		"authentication flow broken",
		"database migration failed",
	}

	vocab := NewVocabulary(docs)
	vec := vocab.Embed("authentication flow")

	if len(vec) != vocab.Size() {
		t.Errorf("vector length = %d, want %d", len(vec), vocab.Size())
	}

	// Vector should be normalized (unit length)
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 0.001 {
		t.Errorf("vector norm = %f, want ~1.0", norm)
	}
}

func TestEmbedEmptyText(t *testing.T) {
	vocab := NewVocabulary([]string{"hello world"})
	vec := vocab.Embed("")
	// Should return zero vector
	for i, v := range vec {
		if v != 0 {
			t.Errorf("vec[%d] = %f, want 0 for empty text", i, v)
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float64
		want float64
	}{
		{"identical", []float64{1, 0, 0}, []float64{1, 0, 0}, 1.0},
		{"orthogonal", []float64{1, 0, 0}, []float64{0, 1, 0}, 0.0},
		{"opposite", []float64{1, 0}, []float64{-1, 0}, -1.0},
		{"empty", []float64{}, []float64{}, 0.0},
		{"zero vector", []float64{0, 0, 0}, []float64{1, 0, 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("CosineSimilarity() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestSimilarDocuments(t *testing.T) {
	docs := []string{
		"authentication login broken error",
		"database migration schema update",
		"authentication session token expired",
		"css styling layout flexbox",
	}

	vocab := NewVocabulary(docs)

	query := vocab.Embed("authentication login session")
	docVecs := make([][]float64, len(docs))
	for i, d := range docs {
		docVecs[i] = vocab.Embed(d)
	}

	// Query about authentication should be most similar to docs 0 and 2
	sim0 := CosineSimilarity(query, docVecs[0])
	sim1 := CosineSimilarity(query, docVecs[1])
	sim2 := CosineSimilarity(query, docVecs[2])
	sim3 := CosineSimilarity(query, docVecs[3])

	if sim0 <= sim1 {
		t.Errorf("auth doc0 sim=%f should be > db doc1 sim=%f", sim0, sim1)
	}
	if sim2 <= sim3 {
		t.Errorf("auth doc2 sim=%f should be > css doc3 sim=%f", sim2, sim3)
	}
}

func TestEncodeDecodeVector(t *testing.T) {
	original := []float64{1.5, -2.3, 0.0, 42.0}
	encoded := EncodeVector(original)
	decoded := DecodeVector(encoded)

	if len(decoded) != len(original) {
		t.Fatalf("decoded length = %d, want %d", len(decoded), len(original))
	}

	for i := range original {
		if math.Abs(decoded[i]-original[i]) > 0.0001 {
			t.Errorf("decoded[%d] = %f, want %f", i, decoded[i], original[i])
		}
	}
}

func TestDecodeVectorEmpty(t *testing.T) {
	decoded := DecodeVector(nil)
	if len(decoded) != 0 {
		t.Errorf("decoded empty = %v, want empty", decoded)
	}
}
