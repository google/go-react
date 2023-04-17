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
	"fmt"
	"testing"

	"github.com/google/go-react/pkg/predictors"
	predictorstesting "github.com/google/go-react/pkg/predictors/testing"
	"github.com/google/go-react/pkg/prompters"
)

func TestRetrier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(*predictorstesting.Fake[PromptData, ParserData])
		assert func(*testing.T, *predictorstesting.Fake[PromptData, ParserData], ParserData, error)
	}{
		{
			name: "no error",
			setup: func(f *predictorstesting.Fake[PromptData, ParserData]) {
				f.Resps = []ParserData{"some response"}
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData, ParserData], resp ParserData, err error) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := len(f.Reqs), 1; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}

				if actual, expected := f.Reqs[0], PromptData(99); actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}

				if actual, expected := resp, ParserData("some response"); actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "hydrate error - no retry",
			setup: func(f *predictorstesting.Fake[PromptData, ParserData]) {
				f.Err = fmt.Errorf("%w: some-error", prompters.ErrHydrate)
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData, ParserData], resp ParserData, err error) {
				if err == nil {
					t.Fatalf("expected error")
				}

				if actual, expected := len(f.Reqs), 1; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
			},
		},
		{
			name: "LLM error - retries",
			setup: func(f *predictorstesting.Fake[PromptData, ParserData]) {
				f.Err = fmt.Errorf("%w: some-error", predictors.ErrLLM)
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData, ParserData], resp ParserData, err error) {
				if err == nil {
					t.Fatalf("expected error")
				}

				if actual, expected := len(f.Reqs), 3; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
			},
		},
		{
			name: "Parse error - retries",
			setup: func(f *predictorstesting.Fake[PromptData, ParserData]) {
				f.Err = fmt.Errorf("%w: some-error", predictors.ErrParse)
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[PromptData, ParserData], resp ParserData, err error) {
				if err == nil {
					t.Fatalf("expected error")
				}

				if actual, expected := len(f.Reqs), 3; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := &predictorstesting.Fake[PromptData, ParserData]{}
			if tc.setup != nil {
				tc.setup(f)
			}

			r := predictors.NewRetrier[PromptData, ParserData](f)
			resp, err := r.Predict(context.Background(), 99)
			tc.assert(t, f, resp, err)
		})
	}
}
