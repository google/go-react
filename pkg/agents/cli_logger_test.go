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
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/go-react/pkg/agents"
	predictorstesting "github.com/google/go-react/pkg/predictors/testing"
)

func TestCLILogger(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		opts   []agents.CLILoggerOption[FinalAnswer]
		req    agents.PromptData[FinalAnswer]
		setup  func(*predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]])
		assert func(*testing.T, *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]], agents.Reasoning[FinalAnswer], error, string)
	}{
		{
			name: "predictor returns Action",
			req:  agents.PromptData[FinalAnswer]{Goal: "some-goal"},
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = []agents.Reasoning[FinalAnswer]{
					{Thought: "some-thought", Action: "some-action", Input: "some-input"},
				}
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]], actualResp agents.Reasoning[FinalAnswer], err error, out string) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := out, "\x1b[32mThought: \"some-thought\" Action: \"some-action\" Input: \"some-input\"\x1b[0m\n"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "predictor returns FinalAnswer",
			req:  agents.PromptData[FinalAnswer]{Goal: "some-goal"},
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = []agents.Reasoning[FinalAnswer]{
					{Thought: "some-thought", FinalAnswer: "some-answer"},
				}
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]], actualResp agents.Reasoning[FinalAnswer], err error, out string) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := out, "\x1b[32mThought: \"some-thought\" FinalAnswer: some-answer\x1b[0m\n"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "passthrough requests and responses",
			req:  agents.PromptData[FinalAnswer]{Goal: "some-goal"},
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = []agents.Reasoning[FinalAnswer]{
					{Thought: "some-thought", Action: "some-action"},
				}
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]], actualResp agents.Reasoning[FinalAnswer], err error, out string) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := actualResp.Action, "some-action"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := len(f.Reqs), 1; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := f.Reqs[0].Goal, "some-goal"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "predictor returns an error",
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Err = errors.New("some error")
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]], actualResp agents.Reasoning[FinalAnswer], err error, out string) {
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			name: "sets a prefix",
			opts: []agents.CLILoggerOption[FinalAnswer]{
				agents.WithCLILoggerPrefix[FinalAnswer]("some-prefix"),
			},
			setup: func(f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]]) {
				f.Resps = []agents.Reasoning[FinalAnswer]{
					{Thought: "some-thought", Action: "some-action", Input: "some-input"},
				}
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[agents.PromptData[FinalAnswer], agents.Reasoning[FinalAnswer]], actualResp agents.Reasoning[FinalAnswer], err error, out string) {
				if actual, expected := out, "\x1b[34m[some-prefix]\x1b[0m \x1b[32mThought: \"some-thought\" Action: \"some-action\" Input: \"some-input\"\x1b[0m\n"; actual != expected {
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
			buf := &bytes.Buffer{}
			logger := agents.NewCLILogger[FinalAnswer](f, buf, tc.opts...)

			resp, err := logger.Predict(context.Background(), tc.req)
			tc.assert(t, f, resp, err, buf.String())
		})
	}
}
