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

package confirmation_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/go-react/pkg/chains/confirmation"
	llmstesting "github.com/google/go-react/pkg/llms/testing"
)

type LLMParams int

func TestConfirmation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		assert func(t *testing.T, result confirmation.Response, err error)
		setup  func(llm *llmstesting.Fake[LLMParams])
	}{
		{
			name: "LLM returns an error",
			setup: func(llm *llmstesting.Fake[LLMParams]) {
				llm.Err = errors.New("some error")
			},
			assert: func(t *testing.T, result confirmation.Response, err error) {
				if err == nil {
					t.Fatal("expected an error")
				}
			},
		},
		{
			name: "LLM returns garbage",
			setup: func(llm *llmstesting.Fake[LLMParams]) {
				llm.AlwaysText = "garbage"
			},
			assert: func(t *testing.T, result confirmation.Response, err error) {
				if err == nil {
					t.Fatal("expected an error")
				}
			},
		},
		{
			name: "LLM returns yes",
			setup: func(llm *llmstesting.Fake[LLMParams]) {
				llm.AlwaysText = `{"result":true}`
			},
			assert: func(t *testing.T, result confirmation.Response, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := result.Result, true; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "LLM returns no",
			setup: func(llm *llmstesting.Fake[LLMParams]) {
				llm.AlwaysText = `{"result":false}`
			},
			assert: func(t *testing.T, result confirmation.Response, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := result.Result, false; actual != expected {
					t.Fatalf("expected %v, got %v", expected, actual)
				}
			},
		},
		{
			name: "LLM user error",
			setup: func(llm *llmstesting.Fake[LLMParams]) {
				llm.AlwaysText = `{"error":"some-user-error"}`
			},
			assert: func(t *testing.T, result confirmation.Response, err error) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := result.Error, "some-user-error"; actual != expected {
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
			if tc.setup != nil {
				tc.setup(llm)
			}

			out := &bytes.Buffer{}
			c := confirmation.New[LLMParams](llm, alwaysEOFReader{}, out)
			result, err := c.Run(context.Background(), "input")

			tc.assert(t, result, err)
		})
	}
}

type alwaysEOFReader struct{}

func (alwaysEOFReader) Read(buffer []byte) (int, error) {
	buffer[0] = 'y'
	buffer[1] = '\n'
	return 2, io.EOF
}
