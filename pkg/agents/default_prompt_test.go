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

package agents_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-react/pkg/agents"
	"github.com/google/go-react/pkg/prompters"
)

func TestDefaultPrompt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		opts   []prompters.Option[agents.PromptData[string]]
		assert func(t *testing.T, result string)
	}{
		{
			name: "default examples",
			assert: func(t *testing.T, result string) {
				if actual, expected := strings.Contains(result, "Example 0"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
				if actual, expected := strings.Contains(result, "Example 1"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
				if actual, expected := strings.Contains(result, "Example 2"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
			},
		},
		{
			name: "custom examples",
			opts: []prompters.Option[agents.PromptData[string]]{
				agents.WithExamples[int, string](agents.PromptDataExample[string]{
					Question: "some fancy example",
				}),
			},
			assert: func(t *testing.T, result string) {
				if actual, expected := strings.Contains(result, "some fancy example"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
			},
		},
		{
			name: "default preamble",
			assert: func(t *testing.T, result string) {
				if actual, expected := strings.Contains(result, "Example 0"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
				if actual, expected := strings.Contains(result, "Example 1"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
				if actual, expected := strings.Contains(result, "Example 2"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
			},
		},
		{
			name: "custom preamble",
			opts: []prompters.Option[agents.PromptData[string]]{
				agents.WithPreamble[int, string]("some fancy preamble"),
			},
			assert: func(t *testing.T, result string) {
				if actual, expected := strings.Contains(result, "some fancy preamble"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
			},
		},
		{
			name: "default rules",
			assert: func(t *testing.T, result string) {
				if actual, expected := strings.Contains(result, "Each thought must follow a plan and should be based on previous thoughts and actions."), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
			},
		},
		{
			name: "custom rules",
			opts: []prompters.Option[agents.PromptData[string]]{
				agents.WithRules[int, string]("some fancy rule"),
			},
			assert: func(t *testing.T, result string) {
				if actual, expected := strings.Contains(result, "some fancy rule"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prompter := agents.NewDefaultPrompt[int, string](0, tc.opts...)
			result, params, err := prompter.Hydrate(context.Background(), agents.PromptData[string]{})
			if err != nil {
				t.Fatal(err)
			}
			if actual, expected := params, 0; actual != expected {
				t.Errorf("got %d, want %d", actual, expected)
			}
			tc.assert(t, result)
		})
	}
}

func TestDefaultPrompt_nonStringType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		opts   []prompters.Option[agents.PromptData[int]]
		assert func(t *testing.T, result string)
	}{
		{
			name: "default examples",
			assert: func(t *testing.T, result string) {
				if actual, expected := strings.Contains(result, "Example 0"), true; actual != expected {
					t.Errorf("got %v, want %v", actual, expected)
				}
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prompter := agents.NewDefaultPrompt[int, int](0, tc.opts...)
			result, params, err := prompter.Hydrate(context.Background(), agents.PromptData[int]{})
			if err != nil {
				t.Fatal(err)
			}
			if actual, expected := params, 0; actual != expected {
				t.Errorf("got %d, want %d", actual, expected)
			}
			tc.assert(t, result)
		})
	}
}
