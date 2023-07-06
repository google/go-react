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
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-react/pkg/prompters"
)

func TestTextTemplate(t *testing.T) {
	t.Parallel()
	p := prompters.NewTextTemplate[map[string]any, int](
		"Hello {{.Name}}",
		99,
		func(p map[string]any) map[string]any {
			p["Name"] = strings.ToUpper(p["Name"].(string))
			return p
		},
	)
	val, params, err := p.Hydrate(context.Background(), map[string]any{
		"Name": "World",
	})
	if err != nil {
		t.Fatal(err)
	}
	if actual, expected := val, "Hello WORLD"; actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
	if actual, expected := params, 99; actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
}

func TestTextTemplate_InvalidTemplate(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	prompters.NewTextTemplate[map[string]any, int]("Hello {{.Name", 99)
}

func TestTextTemplate_InvalidData(t *testing.T) {
	t.Parallel()

	p := prompters.NewTextTemplate[map[string]any, int]("Hello {{.Unknown}}", 99)
	_, _, err := p.Hydrate(context.Background(), nil)
	if actual, expected := errors.Is(err, prompters.ErrHydrate), true; actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
