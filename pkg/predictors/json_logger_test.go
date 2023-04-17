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
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/go-react/pkg/predictors"
	predictorstesting "github.com/google/go-react/pkg/predictors/testing"
)

func TestJSONLogger(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(*predictorstesting.Fake[int, string])
		assert func(*testing.T, *predictorstesting.Fake[int, string], string, error, string)
	}{
		{
			name: "predictor succeeds",
			setup: func(f *predictorstesting.Fake[int, string]) {
				f.Resps = []string{"some response"}
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[int, string], actualResp string, err error, out string) {
				if err != nil {
					t.Fatal(err)
				}
				if actual, expected := actualResp, "some response"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := len(f.Reqs), 1; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := f.Reqs[0], 99; actual != expected {
					t.Fatalf("expected %d, got %d", expected, actual)
				}
				if actual, expected := out, "{\"request\":99}\n{\"response\":\"some response\"}\n"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "predictor returns an error",
			setup: func(f *predictorstesting.Fake[int, string]) {
				f.Err = errors.New("some error")
			},
			assert: func(t *testing.T, f *predictorstesting.Fake[int, string], actualResp string, err error, out string) {
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

			f := &predictorstesting.Fake[int, string]{}
			if tc.setup != nil {
				tc.setup(f)
			}
			buf := &bytes.Buffer{}
			logger := predictors.NewJSONLogger[int, string](f, buf)

			resp, err := logger.Predict(context.Background(), 99)
			tc.assert(t, f, resp, err, buf.String())
		})
	}
}
