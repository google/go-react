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

package agents

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-react/pkg/predictors"
	predictorstesting "github.com/google/go-react/pkg/predictors/testing"
	"github.com/google/go-react/pkg/tools"
)

type FinalAnswer string

func TestReasoningPredictor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		req    PromptData[FinalAnswer]
		setup  func(*predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]])
		assert func(*testing.T, *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], Reasoning[FinalAnswer], error)
	}{
		{
			name: "success - FinalAnswer",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Thought:     "I know the final answer",
					FinalAnswer: "42",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := r.Thought, "I know the final answer"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := r.FinalAnswer, FinalAnswer("42"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := r.Action, ""; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "success - Action",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Thought: "I need to read the columns",
					Action:  "list-columns",
					Input:   "some-table",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := r.Thought, "I need to read the columns"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := r.Action, "list-columns"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := r.Input, "some-table"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := r.FinalAnswer, FinalAnswer(""); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "success - Action has weird casing",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Thought: "I need to read the columns",
					Action:  "list-ColUMNs",
					Input:   "some-table",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := r.Action, "list-columns"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "success - replace all underscores and spacing with snake case",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Thought: "I need to read the columns",
					Action:  "list\t_ ColUMNs",
					Input:   "some-table",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := r.Action, "list-columns"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "missing thought",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Action: "list-columns",
					Input:  "some-table",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if actual, expected := errors.Is(err, predictors.ErrParse), true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "no Action or FinalAnswer",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Thought: "I know the final answer",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if actual, expected := errors.Is(err, predictors.ErrParse), true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "both Action and FinalAnswer",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Thought:     "I know the final answer",
					Action:      "list-columns",
					FinalAnswer: "42",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if actual, expected := errors.Is(err, predictors.ErrParse), true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "Action references an unknown tool",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, Reasoning[FinalAnswer]{
					Thought: "I know the final answer",
					Action:  "Unknown",
				})
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if actual, expected := errors.Is(err, ErrInvalidTool), true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "passthrough error",
			setup: func(f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]) {
				f.Err = errors.New("some error")
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]], r Reasoning[FinalAnswer], err error) {
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := &predictorstesting.Fake[PromptData[FinalAnswer], Reasoning[FinalAnswer]]{}
			if tc.setup != nil {
				tc.setup(f)
			}
			tools := []tools.Tool{
				// NOTE: The tool shouldn't actually be used, just the name. We use
				// weird casing here to make sure everything is normalized.
				{Name: "list-coluMNS"},
			}

			p := newReasoningPredictor[FinalAnswer](f, tools)
			result, err := p.Predict(context.Background(), tc.req)
			tc.assert(t, f, result, err)
		})
	}
}
