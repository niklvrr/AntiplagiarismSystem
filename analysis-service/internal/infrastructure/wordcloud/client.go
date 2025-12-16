package wordcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"go.uber.org/zap"
)

type Client struct {
	baseURL string
	logger  *zap.Logger
}

func NewClient(logger *zap.Logger) *Client {
	return &Client{
		baseURL: "https://quickchart.io/wordcloud",
		logger:  logger,
	}
}

func (c *Client) GenerateWordCloud(ctx context.Context, fileContent []byte) (string, error) {
	c.logger.Info("generating word cloud",
		zap.Int("file_size", len(fileContent)))

	words := c.extractWords(fileContent)
	wordFreq := c.countWordFrequency(words)

	c.logger.Debug("word frequency calculated",
		zap.Int("unique_words", len(wordFreq)),
		zap.Int("total_words", len(words)))

	text := strings.Join(words, " ")
	
	config := map[string]interface{}{
		"type":    "wordCloud",
		"data":    wordFreq,
		"options": map[string]interface{}{
			"width":   800,
			"height":  600,
			"colors":  []string{"#1f77b4", "#ff7f0e", "#2ca02c", "#d62728", "#9467bd", "#8c564b"},
			"fontFamily": "Arial",
			"scale":   "sqrt",
		},
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		c.logger.Error("failed to marshal word cloud config", zap.Error(err))
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	chartURL := fmt.Sprintf("%s?text=%s&c=%s", c.baseURL, url.QueryEscape(text), url.QueryEscape(string(configJSON)))

	c.logger.Info("word cloud generated",
		zap.String("url", chartURL))

	return chartURL, nil
}

func (c *Client) extractWords(content []byte) []string {
	text := string(content)
	text = strings.ToLower(text)

	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 2 {
				words = append(words, currentWord.String())
			}
			currentWord.Reset()
		}
	}

	if currentWord.Len() > 2 {
		words = append(words, currentWord.String())
	}

	return words
}

func (c *Client) countWordFrequency(words []string) map[string]int {
	freq := make(map[string]int)
	stopWords := map[string]bool{
		"the": true, "be": true, "to": true, "of": true, "and": true,
		"a": true, "in": true, "that": true, "have": true, "i": true,
		"it": true, "for": true, "not": true, "on": true, "with": true,
		"he": true, "as": true, "you": true, "do": true, "at": true,
		"this": true, "but": true, "his": true, "by": true, "from": true,
		"they": true, "we": true, "say": true, "her": true, "she": true,
		"or": true, "an": true, "will": true, "my": true, "one": true,
		"all": true, "would": true, "there": true, "their": true,
	}

	for _, word := range words {
		if !stopWords[word] && len(word) > 2 {
			freq[word]++
		}
	}

	return freq
}

func (c *Client) GenerateWordCloudImage(ctx context.Context, fileContent []byte) ([]byte, error) {
	chartURL, err := c.GenerateWordCloud(ctx, fileContent)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("fetching word cloud image", zap.String("url", chartURL))

	req, err := http.NewRequestWithContext(ctx, "GET", chartURL, nil)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("failed to fetch word cloud image", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code",
			zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read image data", zap.Error(err))
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	c.logger.Info("word cloud image fetched",
		zap.Int("image_size", len(imageData)))

	return imageData, nil
}

