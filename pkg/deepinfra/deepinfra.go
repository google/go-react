package deepinfra

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-react/pkg/llms/vertex"
)

type DeepInfraLLM struct {
	apiKey string
}

func NewDeepInfraLLM(apiKey string) *DeepInfraLLM {
	return &DeepInfraLLM{apiKey: apiKey}
}

func (d *DeepInfraLLM) Generate(ctx context.Context, prompt string, params vertex.Params) (string, error) {
	url := "https://api.deepinfra.com/v1/openai/chat/completions"

	requestBody := map[string]interface{}{
		"model": params.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  params.MaxTokens,
		"temperature": params.Temperature,
		"top_k":       params.TopK,
		"top_p":       params.TopP,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no choices returned in the response")
	}

	return response.Choices[0].Message.Content, nil
}
