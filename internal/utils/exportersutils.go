package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Copyright 2026 oyama forked cotsom

// llm helped here oops
func ParseExportersType(body io.ReadCloser) (string, error) {
	defer body.Close()

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", fmt.Errorf("не удалось распарсить HTML: %w", err)
	}

	title := doc.Find("header h1").First().Text()
	title = strings.TrimSpace(title)

	if title == "" {
		// fallback — ищем любой h1
		title = doc.Find("h1").First().Text()
		title = strings.TrimSpace(title)
	}

	return title, nil
}

func UnmarshallJsonString(jsonString string) (map[string]interface{}, error) {
	if jsonString == "" {
		return nil, fmt.Errorf("json string is empty")
	}

	jsonString = strings.TrimSpace(jsonString)
	if jsonString == "" {
		return nil, fmt.Errorf("json string is empty after trim")
	}

	var result map[string]interface{}

	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("JSON root is not an object (got null or array/primitive)")
	}

	return result, nil
}

func ExportersExtractCmdline(data map[string]interface{}) (string, bool) {
	if data == nil {
		return "", false
	}

	cmdlineIface, ok := data["cmdline"]
	if !ok || cmdlineIface == nil {
		return "", false
	}

	switch v := cmdlineIface.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			return v, true
		}
	case []interface{}:
		var sb strings.Builder
		for _, arg := range v {
			if arg != nil {
				sb.WriteString(fmt.Sprintf("%v ", arg))
			}
		}
		result := strings.TrimSpace(sb.String())
		if result != "" {
			return result, true
		}
	default:
		s := fmt.Sprintf("%v", v)
		if strings.TrimSpace(s) != "" {
			return s, true
		}
	}

	return "", false
}
