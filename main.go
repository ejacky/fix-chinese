package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Request struct {
	Text string `json:"text"`
}

func main() {
	r := gin.Default()

	r.POST("/api/correct", func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		result, err := callLLM(req.Text)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.Data(200, "application/json", result)
	})

	r.Run(":8080")
}

func callLLM(input string) ([]byte, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")

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
		return nil, err
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	content := res["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	return []byte(content), nil
}
