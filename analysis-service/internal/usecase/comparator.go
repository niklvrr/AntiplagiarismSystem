package usecase

import (
	"context"
	"strings"
	"unicode"
)

type TextComparator struct{}

func NewTextComparator() *TextComparator {
	return &TextComparator{}
}

func (c *TextComparator) CompareFiles(ctx context.Context, file1, file2 []byte) (float64, error) {
	text1 := string(file1)
	text2 := string(file2)

	text1 = normalizeText(text1)
	text2 = normalizeText(text2)

	similarity := calculateSimilarity(text1, text2)

	return similarity, nil
}

func normalizeText(text string) string {
	var builder strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}

func calculateSimilarity(text1, text2 string) float64 {
	if len(text1) == 0 && len(text2) == 0 {
		return 100.0
	}
	if len(text1) == 0 || len(text2) == 0 {
		return 0.0
	}

	n := 3
	ngrams1 := createNGrams(text1, n)
	ngrams2 := createNGrams(text2, n)

	intersection := 0
	for ngram := range ngrams1 {
		if ngrams2[ngram] {
			intersection++
		}
	}

	union := len(ngrams1) + len(ngrams2) - intersection
	if union == 0 {
		return 0.0
	}

	return (float64(intersection) / float64(union)) * 100.0
}

func createNGrams(text string, n int) map[string]bool {
	ngrams := make(map[string]bool)
	for i := 0; i <= len(text)-n; i++ {
		ngram := text[i : i+n]
		ngrams[ngram] = true
	}
	return ngrams
}
