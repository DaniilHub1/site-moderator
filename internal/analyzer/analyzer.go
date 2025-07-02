package analyzer

import (
	"encoding/json"
	"os"
	"strings"
)

type KeywordMap map[string][]string

var keywords KeywordMap

var safeContext = []string{
	"жанр", "фильм", "сериал", "режиссер", "смотреть", "в ролях", "описание", "hd", "онлайн",
}

func LoadKeywords(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &keywords)
}

type MatchResult struct {
	Category string
	Word     string
	Context  string
	IsSuspected   bool
}

func CheckText(content string) (*MatchResult, bool) {
	lower := strings.ToLower(content)
	for category, list := range keywords {
		for _, word := range list {
			if idx := strings.Index(lower, word); idx != -1 {
				start := max(0, idx-50)
				end := min(len(lower), idx+50)
				context := lower[start:end]

				suspected := true
				for _, safe := range safeContext {
					if strings.Contains(context, safe) {
						suspected = false
						break
					}
				}

				return &MatchResult{
					Category:    category,
					Word:        word,
					Context:     context,
					IsSuspected: suspected,
				}, true
			}
		}
	}
	return nil, false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
