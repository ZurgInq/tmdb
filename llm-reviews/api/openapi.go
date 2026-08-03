package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type ChatRequest struct {
	Messages []Message `json:"messages"`

	Model             string          `json:"model"`
	ResponseFormat    json.RawMessage `json:"response_format"`
	Stream            bool            `json:"stream"`
	Mode              string          `json:"mode,omitempty"`
	Temperature       float32         `json:"temperature,omitempty"`
	TopK              int             `json:"top_k,omitempty"`
	TopP              float32         `json:"top_p,omitempty"`
	MinP              float32         `json:"min_p,omitempty"`
	PresencePenalty   float32         `json:"presence_penalty,omitempty"`
	RepetitionPenalty float32         `json:"repetition_penalty,omitempty"`

	// Reasoning Format
	ReasoningFormat  string `json:"reasoning_format,omitempty"` // parsed, raw, hidden
	IncludeReasoning bool   `json:"include_reasoning,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Client struct {
	ApiKey              string
	ApiURL              string
	Client              *http.Client
	MaxRetries          int
	ChatRequestModifier func(*ChatRequest) any
}

func (c *Client) SendRequest(body []byte) (*ChatResponse, error) {
	backoff := 2 * time.Second

	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.ApiURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		if len(c.ApiKey) > 0 {
			req.Header.Set("Authorization", "Bearer "+c.ApiKey)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.Client.Do(req)
		if err != nil {
			if attempt == c.MaxRetries {
				return nil, err
			}

			fmt.Printf("network error: %v, retry after %v\n", err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		// Успех
		if resp.StatusCode == http.StatusOK {
			var chatResp ChatResponse
			if err := json.Unmarshal(respBody, &chatResp); err != nil {
				return nil, err
			}
			return &chatResp, nil
		}

		fmt.Println(string(respBody))
		// Rate limit
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := backoff

			// Если сервер прислал Retry-After — используем его
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if sec, err := strconv.Atoi(retryAfter); err == nil {
					delay = time.Duration(sec) * time.Second
				}
			}

			fmt.Printf("Rate limit exceeded (attempt %d/%d), sleep %v\n",
				attempt, c.MaxRetries, delay)

			time.Sleep(delay)
			backoff *= 2
			continue
		}

		// Временные ошибки сервера
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			fmt.Printf("Server error %d, retry after %v\n",
				resp.StatusCode, backoff)

			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		// Остальные ошибки не повторяем
		return nil, fmt.Errorf("status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return nil, fmt.Errorf("maximum retries exceeded")
}

func (client *Client) MarshalChatRequest(chatRequest ChatRequest) ([]byte, error) {
	if client.ChatRequestModifier != nil {
		return json.Marshal(client.ChatRequestModifier(&chatRequest))
	}
	return json.Marshal(chatRequest)
}

func (client *Client) SendInstruct(content string) (string, error) {
	body, err := client.MarshalChatRequest(ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: content,
			},
		},
		Stream:           false,
		IncludeReasoning: false,
	})

	if err != nil {
		return "", err
	}

	chatResp, err := client.SendRequest(body)
	if err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		fmt.Println("No choices returned")
		return "", nil
	}

	return chatResp.Choices[0].Message.Content, nil
}
