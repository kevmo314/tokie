package tokie

import (
	"bufio"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"go.elara.ws/pcre"
)

type bytePairEncoding struct {
	tokenForRank [][]byte
	rankForToken map[string]uint32
}

func NewBytePairEncoderFromTokenizerModel(r io.Reader) (*bytePairEncoding, error) {
	scanner := bufio.NewScanner(r)
	vocab := [][]byte{}
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		b, err := base64.StdEncoding.DecodeString(tokens[0])
		if err != nil {
			return nil, err
		}
		vocab = append(vocab, b)
	}
	return NewBytePairEncoder(vocab), nil
}

func NewBytePairEncoder(vocab [][]byte) *bytePairEncoding {
	rankForToken := map[string]uint32{}
	for i, token := range vocab {
		rankForToken[string(token)] = uint32(i)
	}
	return &bytePairEncoding{
		tokenForRank: vocab,
		rankForToken: rankForToken,
	}
}

var errInvalidToken = errors.New("invalid token")

func (bpe *bytePairEncoding) EncodeSplitting(ts []int, bs []byte, re *pcre.Regexp) (int, error) {
	i := 0
	for _, s := range re.FindAll(bs, -1) {
		j, err := bpe.Encode(ts[i:], s)
		if err != nil {
			return 0, err
		}
		i += j
	}
	return i, nil
}

func (bpe *bytePairEncoding) Encode(ts []int, bs []byte) (int, error) {
	// handle some edge cases
	if len(bs) == 0 {
		return 0, nil
	} else if len(bs) == 1 {
		ts[0] = int(bs[0])
		return 1, nil
	}

	// the length of tokens may be shorter when output but we need some working space
	if len(ts) < len(bs) {
		return 0, io.ErrShortWrite
	}

	// ranks is the rank that would result if the token at i was merged
	// with the token at next[i]. this is cached to improve performance as it
	// makes the minimum rank finding loop faster
	ranks := make([]uint32, len(bs))
	next := make([]int, len(bs))
	prev := make([]int, len(bs)+1)
	prev[0] = -1
	minRankIndex := 0
	for i := 0; i < len(bs); i++ {
		if i < len(bs)-1 {
			rank, ok := bpe.rankForToken[string(bs[i:i+2])]
			if ok {
				ranks[i] = rank
			} else {
				ranks[i] = ^uint32(0)
			}
		} else {
			ranks[i] = ^uint32(0)
		}
		if ranks[i] < ranks[minRankIndex] {
			minRankIndex = i
		}
		prev[i+1] = i
		next[i] = i + 1
	}

	for {
		if ranks[minRankIndex] == ^uint32(0) {
			// we're done, nothing left to merge
			break
		}

		// merge with next token
		next[minRankIndex] = next[next[minRankIndex]]
		prev[next[minRankIndex]] = minRankIndex

		// update the rank of the current pair.
		if next[minRankIndex] < len(bs) {
			rank, ok := bpe.rankForToken[string(bs[minRankIndex:next[next[minRankIndex]]])]
			if ok {
				ranks[minRankIndex] = rank
			} else {
				ranks[minRankIndex] = ^uint32(0)
			}
		} else {
			ranks[minRankIndex] = ^uint32(0)
		}

		// update the rank of the previous pair
		if prev[minRankIndex] >= 0 {
			rank, ok := bpe.rankForToken[string(bs[prev[minRankIndex]:next[minRankIndex]])]
			if ok {
				ranks[prev[minRankIndex]] = rank
			} else {
				ranks[prev[minRankIndex]] = ^uint32(0)
			}
		}

		// find the next minimum rank
		for j := 0; j < len(bs); j = next[j] {
			if ranks[j] < ranks[minRankIndex] {
				minRankIndex = j
			}
		}
	}
	j := 0
	for i := 0; i < len(bs); i, j = next[i], j+1 {
		rank, ok := bpe.rankForToken[string(bs[i:next[i]])]
		if ok {
			ts[j] = int(rank)
		} else {
			panic("invariant violated")
		}
	}
	return j, nil
}

func (bpe *bytePairEncoding) Decode(bs []byte, ts []int) (int, error) {
	i := 0
	for _, t := range ts {
		if t >= len(bpe.tokenForRank) {
			return 0, errInvalidToken
		}
		tb := bpe.tokenForRank[t]
		if copy(bs[i:], tb) < len(tb) {
			return 0, io.ErrShortWrite
		}
		i += len(tb)
	}
	return i, nil
}
