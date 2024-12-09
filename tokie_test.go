package tokie

import (
	"log"
	"os"
	"testing"

	"go.elara.ws/pcre"
)

func TestEncode(t *testing.T) {
	f, err := os.Open("tokenizer.model")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	re := pcre.MustCompile(`(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`)

	bpe, err := NewBytePairEncoderFromTokenizerModel(f)
	if err != nil {
		t.Fatal(err)
	}

	tokens := make([]int, 32)

	if _, err := bpe.EncodeSplitting(tokens, []byte("do you know the muffin man? the muffin man? the muffin man who lives down the road"), re); err != nil {
		t.Fatal(err)
	}
	log.Printf("%v", tokens)
}

func TestBulkEncode(t *testing.T) {
	f, err := os.Open("tokenizer.model")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	g, err := os.ReadFile("world192.txt")
	if err != nil {
		t.Fatal(err)
	}

	re := pcre.MustCompile(`(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`)

	bpe, err := NewBytePairEncoderFromTokenizerModel(f)
	if err != nil {
		t.Fatal(err)
	}

	tokens := make([]int, len(g))

	n, err := bpe.EncodeSplitting(tokens, g, re)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("tokens: %d", n)
}

func BenchmarkEncode(b *testing.B) {
	f, err := os.Open("tokenizer.model")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	bpe, err := NewBytePairEncoderFromTokenizerModel(f)
	if err != nil {
		b.Fatal(err)
	}

	tokens := make([]int, 32)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := bpe.Encode(tokens, []byte("nondifferentiable")); err != nil {
			b.Fatal(err)
		}
	}
}
