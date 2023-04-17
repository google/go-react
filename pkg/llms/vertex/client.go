// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vertex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-react/pkg/llms"
	oauth2 "golang.org/x/oauth2/google"
)

// Params are the parameters for a request.
type Params struct {
	Model       string
	MaxTokens   int
	Temperature float64
	TopK        int
	TopP        float64
}

// New returns a new Google Palm LLM.
func New(ctx context.Context, apiEndpoint, projectID string) (llms.LLM[Params], error) {
	c, err := oauth2.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}

	return client{
		projectID:   projectID,
		apiEndpoint: apiEndpoint,
		client:      c,
	}, nil
}

// NewWithKey returns a new Google Palm LLM.
func NewWithKey(key, apiEndpoint, projectID string) llms.LLM[Params] {
	return client{
		key:         key,
		projectID:   projectID,
		apiEndpoint: apiEndpoint,
		client:      http.DefaultClient,
	}
}

type client struct {
	key         string
	projectID   string
	apiEndpoint string
	client      *http.Client
}

// Generate implements llms.LLM.
func (c client) Generate(ctx context.Context, prompt string, params Params) (string, error) {
	// Check to see if the params are empty, if so, set some defaults.
	if params == (Params{}) {
		params.Model = "text-bison@001"
		params.MaxTokens = 64
		params.Temperature = 0.2
		params.TopK = 40
		params.TopP = 0.8
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"https://%s/v1/projects/%s/locations/us-central1/publishers/google/models/%s:predict",
			c.apiEndpoint,
			c.projectID,
			params.Model,
		),
		strings.NewReader(fmt.Sprintf(`
{
  "instances": [
    { "content": %q }
  ],
  "parameters": {
	  "temperature": %f,
    "maxOutputTokens": %d,
		"topK": %d,
		"topP": %f
  }
}`, prompt, params.Temperature, params.MaxTokens, params.TopK, params.TopP)),
	)

	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	if c.key != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.key))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %v", err)
		}
		return "", fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, data)
	}

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(r.Predictions) == 0 {
		return "", fmt.Errorf("no predictions returned")
	}

	return r.Predictions[0].Content, nil
}

type response struct {
	Predictions []struct {
		Content string `json:"content"`
	} `json:"predictions"`
}
