package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Request struct {
	Text string `json:"text"`
}

type LLMResult struct {
	Correction  string   `json:"correction"`
	Explanation string   `json:"explanation"`
	Natural     []string `json:"natural"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func main() {
	r := gin.Default()

	// Serve the single-page frontend.
	r.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})
	r.GET("/index.html", func(c *gin.Context) {
		c.File("index.html")
	})

	r.POST("/api/correct", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		result, status, err := callLLM(req.Text)
		if err != nil {
			if status < 400 || status > 599 {
				status = 500
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, result)
	})

	r.Run(":8080")
}

func callLLM(input string) (LLMResult, int, error) {
	var out LLMResult

	apiKey := os.Getenv("OPENAI_API_KEY")
	if strings.TrimSpace(apiKey) == "" {
		return out, http.StatusUnauthorized, errors.New("missing OPENAI_API_KEY")
	}

	payload := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are a professional Chinese teacher helping English-speaking students.

Your job:
- Correct the student's Chinese sentence
- Explain mistakes clearly in simple English
- Provide more natural alternatives used by native speakers

Rules:
1. Be friendly and encouraging
2. Keep explanations simple
3. Always respond in JSON format:
{
  "correction": "...",
  "explanation": "...",
  "natural": ["...", "..."]
}`,
			},
			{
				"role":    "user",
				"content": "Student sentence:\n" + input,
			},
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.gptsapi.net/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return out, http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, http.StatusBadGateway, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Try to surface upstream error message if possible.
		var parsed openAIChatResponse
		_ = json.Unmarshal(respBody, &parsed)
		msg := "upstream LLM request failed"
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			msg = parsed.Error.Message
		} else {
			// Fallback: try generic { "error": { "message": "..." } }.
			var generic map[string]any
			if json.Unmarshal(respBody, &generic) == nil {
				if e, ok := generic["error"].(map[string]any); ok {
					if m, ok := e["message"].(string); ok && strings.TrimSpace(m) != "" {
						msg = m
					}
				}
			}
		}
		return out, resp.StatusCode, errors.New(msg)
	}

	// Parse upstream OpenAI-compatible response.
	var res openAIChatResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return out, http.StatusBadGateway, err
	}
	if len(res.Choices) == 0 {
		return out, http.StatusBadGateway, errors.New("upstream response missing choices")
	}
	content := strings.TrimSpace(res.Choices[0].Message.Content)
	if content == "" {
		return out, http.StatusBadGateway, errors.New("upstream response missing message content")
	}

	// The LLM returns JSON *as a string* in content; parse it into a real JSON object.
	jsonText := extractJSON(content)
	if err := json.Unmarshal([]byte(jsonText), &out); err != nil {
		return out, http.StatusBadGateway, err
	}
	return out, http.StatusOK, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	// If the model wrapped JSON in markdown code fences, attempt to grab the raw JSON block.
	// This keeps parsing robust even when the model isn't perfectly formatted.
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
