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

package prompters_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-react/pkg/prompters"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		template     string
		promptParams map[string]any
		assert       func(*testing.T, string, int, error, string)
	}{
		{
			name:         "passthrough output",
			template:     "{{.foo}}",
			promptParams: map[string]any{"foo": "bar"},
			assert: func(t *testing.T, output string, params int, err error, loggedOutput string) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := output, "bar"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
			},
		},
		{
			name:         "passthrough error",
			template:     "{{.unknown}}",
			promptParams: map[string]any{},
			assert: func(t *testing.T, output string, params int, err error, loggedOutput string) {
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			name:         "writes hydrated prompt",
			template:     "{{.foo}}",
			promptParams: map[string]any{"foo": "bar"},
			assert: func(t *testing.T, output string, params int, err error, loggedOutput string) {
				if err != nil {
					t.Fatal(err)
				}

				if actual, expected := loggedOutput, "bar\n"; actual != expected {
					t.Fatalf("expected %q, got %q", expected, actual)
				}
				if actual, expected := params, 99; actual != expected {
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

			p := prompters.NewTextTemplate[map[string]any, int]("{{.foo}}", 99)
			buf := &bytes.Buffer{}
			l := prompters.NewLogger[map[string]any](p, buf)

			output, params, err := l.Hydrate(context.Background(), tc.promptParams)
			tc.assert(t, output, params, err, buf.String())
		})
	}
}
