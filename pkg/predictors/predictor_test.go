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

package predictors_test

import (
	"context"
	"errors"
	"testing"

	llmstesting "github.com/google/go-react/pkg/llms/testing"
	parserstesting "github.com/google/go-react/pkg/parsers/testing"
	"github.com/google/go-react/pkg/predictors"
	"github.com/google/go-react/pkg/prompters"
	prompterstesting "github.com/google/go-react/pkg/prompters/testing"
)

type LLMParams int
type PromptData int
type ParserData string

func TestPredict(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(*llmstesting.Fake[LLMParams], *prompterstesting.Fake[PromptData, LLMParams], *parserstesting.Fake[ParserData])
		assert func(t *testing.T, resp ParserData, err error, m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData])
	}{
		{
			name: "success",
			setup: func(m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				p.HydrateF = func(context.Context, PromptData) (string, LLMParams, error) {
					return "some-output", 0, nil
				}
				m.AlwaysText = "some-llm-output"
				parser.ParseF = func(intput string) (ParserData, error) {
					return "some-parsed-output", nil
				}
			},
			assert: func(t *testing.T, resp ParserData, err error, m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := p.HydrateData[0], PromptData(1); actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := m.Prompts[0], "some-output"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := parser.Datas[0], "some-llm-output"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := resp, ParserData("some-parsed-output"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "hydrating prompt fails",
			setup: func(m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				p.HydrateF = func(context.Context, PromptData) (string, LLMParams, error) {
					return "", 0, errors.New("some-error")
				}
			},
			assert: func(t *testing.T, resp ParserData, err error, m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				if actual, expected := errors.Is(err, prompters.ErrHydrate), true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "LLM prediction fails",
			setup: func(m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				m.Err = errors.New("some-error")
			},
			assert: func(t *testing.T, resp ParserData, err error, m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				if actual, expected := errors.Is(err, predictors.ErrLLM), true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "parsing LLM response fails",
			setup: func(m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				parser.ParseF = func(intput string) (ParserData, error) {
					return "", errors.New("some-error")
				}
			},
			assert: func(t *testing.T, resp ParserData, err error, m *llmstesting.Fake[LLMParams], p *prompterstesting.Fake[PromptData, LLMParams], parser *parserstesting.Fake[ParserData]) {
				if actual, expected := errors.Is(err, predictors.ErrParse), true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			llm := &llmstesting.Fake[LLMParams]{}
			prompter := &prompterstesting.Fake[PromptData, LLMParams]{}
			parser := &parserstesting.Fake[ParserData]{}

			if tc.setup != nil {
				tc.setup(llm, prompter, parser)
			}

			predictor := predictors.New[PromptData, ParserData, LLMParams](llm, prompter, parser)
			reps, err := predictor.Predict(context.Background(), 1)
			tc.assert(t, reps, err, llm, prompter, parser)
		})
	}
}
