package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type Request struct {
	Text string `json:"text"`
	Lang string `json:"lang"`
}

type LLMResult struct {
	Correction  string   `json:"correction"`
	Explanation string   `json:"explanation"`
	Natural     []string `json:"natural"`
	HSKLevel    string   `json:"hsk_level"`
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

func Handler(w http.ResponseWriter, r *http.Request) {
	// ✅ 只允许 POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// ✅ 限制 body 大小（防止滥用）
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid input", 400)
		return
	}

	// ✅ 限制输入长度（关键）
	if len([]rune(req.Text)) == 0 || len([]rune(req.Text)) > 200 {
		http.Error(w, "text too long or empty", 400)
		return
	}

	result, status, err := callLLM(req.Text, req.Lang)
	if err != nil {
		if status < 400 || status > 599 {
			status = 500
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func callLLM(input, lang string) (LLMResult, int, error) {
	var out LLMResult

	apiKey := os.Getenv("OPENAI_API_KEY")
	if strings.TrimSpace(apiKey) == "" {
		return out, http.StatusUnauthorized, errors.New("missing OPENAI_API_KEY")
	}

	systemPrompt := `You are a professional Chinese teacher helping English-speaking students.

Your job:
- Correct the student's Chinese sentence
- Explain mistakes clearly in simple English
- Provide more natural alternatives used by native speakers
- Estimate the HSK level (1-6) of the original sentence based on vocabulary and grammar used

Rules:
1. Be friendly and encouraging
2. Keep explanations simple
3. Always respond in JSON format:
{
  "correction": "...",
  "explanation": "...",
  "natural": ["...", "..."],
  "hsk_level": "3"
}
If the output is not valid JSON, fix it before responding.`

	if lang == "vi" {
		systemPrompt = `You are a professional Chinese teacher helping Vietnamese students.

Your job:
- Correct the student's Chinese sentence
- Explain mistakes clearly in simple Vietnamese
- Provide more natural alternatives used by native speakers, with brief Vietnamese explanations if helpful
- Estimate the HSK level (1-6) of the original sentence based on vocabulary and grammar used

Rules:
1. Be friendly and encouraging
2. Keep explanations simple and use everyday Vietnamese
3. Always respond in JSON format:
{
  "correction": "...",
  "explanation": "...",
  "natural": ["...", "..."],
  "hsk_level": "3"
}
If the output is not valid JSON, fix it before responding.`
	}

	payload := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
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
		var parsed openAIChatResponse
		_ = json.Unmarshal(respBody, &parsed)

		msg := "upstream LLM request failed"
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			msg = parsed.Error.Message
		}
		return out, resp.StatusCode, errors.New(msg)
	}

	var res openAIChatResponse
	if err := json.Unmarshal(respBody, &res); err != nil {
		return out, http.StatusBadGateway, err
	}
	if len(res.Choices) == 0 {
		return out, http.StatusBadGateway, errors.New("no choices")
	}

	content := strings.TrimSpace(res.Choices[0].Message.Content)
	jsonText := extractJSON(content)

	if err := json.Unmarshal([]byte(jsonText), &out); err != nil {
		return out, http.StatusBadGateway, err
	}

	return out, http.StatusOK, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
