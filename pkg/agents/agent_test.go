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
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-react/pkg/agents"
	predictorstesting "github.com/google/go-react/pkg/predictors/testing"
	"github.com/google/go-react/pkg/tools"
)

type FinalAnswer string

type setupF func(
	*predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]],
)

func TestAgent_Run(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		goal   string
		setup  setupF
		assert func(
			*testing.T,
			FinalAnswer,
			error,
			*predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]],
		)
	}{
		{
			name: "chains together iterations",
			goal: "some goal",
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = append(
					f.Resps,
					agents.Reasoning[FinalAnswer]{
						Thought: "some-thought",
						// Use strange casing to ensure we are properly normalizing tool
						// names.
						Action: "foo_ tOOL",
						Input:  "some-input",
					},
					agents.Reasoning[FinalAnswer]{
						Thought:     "some-thought",
						FinalAnswer: "some-final-answer",
					},
				)
			},
			assert: func(t *testing.T, answer FinalAnswer, err error, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := len(f.Reqs), 2; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}

				// First iteration.
				if actual, expected := f.Reqs[0].Goal, "some goal"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := len(f.Reqs[0].Tools), 2; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := len(f.Reqs[0].Chains), 0; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}

				// Second iteration.
				if actual, expected := f.Reqs[1].Goal, "some goal"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := len(f.Reqs[1].Tools), 2; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := len(f.Reqs[1].Chains), 1; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := f.Reqs[1].Chains[0].Observation, "running tool foo_TOOL: some-input"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}

				if actual, expected := answer, FinalAnswer("some-final-answer"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "returns FinalAnswer",
			goal: "some goal",
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = append(f.Resps, agents.Reasoning[FinalAnswer]{
					Thought:     "some-thought",
					FinalAnswer: "some-final-answer",
				})
			},
			assert: func(t *testing.T, answer FinalAnswer, err error, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := answer, FinalAnswer("some-final-answer"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "empty goal",
			goal: "",
			assert: func(t *testing.T, answer FinalAnswer, err error, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			// NOTE: This test asserts that the necessary predictors are chained together.
			name: "retries on invalid Reasoning[FinalAnswer]",
			goal: "some goal",
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = append(
					f.Resps,
					agents.Reasoning[FinalAnswer]{
						Thought: "",
						Action:  "foo-tool",
					},
					agents.Reasoning[FinalAnswer]{
						Thought:     "some-thought",
						FinalAnswer: "some-final-answer",
					},
				)
			},
			assert: func(t *testing.T, answer FinalAnswer, err error, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := len(f.Reqs), 2; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := answer, FinalAnswer("some-final-answer"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "retries on invalid tool",
			goal: "some goal",
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = append(
					f.Resps,
					agents.Reasoning[FinalAnswer]{
						Thought: "some-thought",
						Action:  "invalid-tool",
					},
					agents.Reasoning[FinalAnswer]{
						Thought:     "some-thought",
						FinalAnswer: "some-final-answer",
					},
				)
			},
			assert: func(t *testing.T, answer FinalAnswer, err error, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := len(f.Reqs), 2; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := len(f.Reqs[1].Chains), 1; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := f.Reqs[1].Chains[0].Observation, `ERROR: invalid tool: "invalid-tool"`; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := answer, FinalAnswer("some-final-answer"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "predictor returns an error",
			goal: "some goal",
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Err = errors.New("some error")
			},
			assert: func(t *testing.T, answer FinalAnswer, err error, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			name: "tool returns an error",
			goal: "some goal",
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = append(
					f.Resps,
					agents.Reasoning[FinalAnswer]{
						Thought: "some-thought",
						Action:  "always-errors",
					},
					agents.Reasoning[FinalAnswer]{
						Thought:     "some-thought",
						FinalAnswer: "some-final-answer",
					},
				)
			},
			assert: func(t *testing.T, answer FinalAnswer, err error, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := len(f.Reqs), 2; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := len(f.Reqs[1].Chains), 1; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := f.Reqs[1].Chains[0].Observation, "ERROR: some error"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := answer, FinalAnswer("some-final-answer"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := &predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]{}
			if tc.setup != nil {
				tc.setup(f)
			}

			// Use strange casing to ensure we properly normalize it.
			r := agents.NewAgent[FinalAnswer](f, buildFakeTool("foo_TOOL"), buildErrorTool())

			answer, err := r.Run(context.Background(), tc.goal)

			tc.assert(t, answer, err, f)
		})
	}
}

func TestAgent_invalidToolSetup(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		toolSet []tools.Tool
	}{
		{
			name:    "empty tool set",
			toolSet: nil,
		},
		{
			name:    "invalid tool",
			toolSet: []tools.Tool{buildInvalidTool()},
		},
		{
			name: "redundant tool names",
			toolSet: []tools.Tool{
				buildFakeTool("foo"),
				buildFakeTool("foo"),
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic")
				}
			}()

			predictor := &predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]{}
			agents.NewAgent[FinalAnswer](predictor, tc.toolSet...)
		})
	}
}

func buildInvalidTool() tools.Tool {
	return tools.Tool{
		Name:        "",
		Description: "",
	}
}

func buildFakeTool(name string) tools.Tool {
	return tools.Tool{
		Name:        name,
		Description: "some-description",
		Args:        []string{"arg"},
		Run: func(ctx context.Context, input string) (any, error) {
			return fmt.Sprintf("running tool %s: %s", name, input), nil
		},
	}
}

func buildErrorTool() tools.Tool {
	return tools.Tool{
		Name:        "always-errors",
		Description: "some-description",
		Args:        []string{"arg"},
		Run: func(context.Context, string) (any, error) {
			return "", errors.New("some error")
		},
	}
}
