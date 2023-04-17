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

package parsers_test

import (
	"testing"

	"github.com/google/go-react/pkg/parsers"
)

func TestJSONParser(t *testing.T) {
	t.Parallel()
	p := parsers.NewJSONParser[map[string]int]()
	val, err := p.Parse(`
{
	"a":
	1
}
extra garbage`)

	if err != nil {
		t.Fatal(err)
	}

	if actual, expected := len(val), 1; actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
	if actual, expected := val["a"], 1; actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
}

func TestJSONParser_invalid(t *testing.T) {
	t.Parallel()
	p := parsers.NewJSONParser[map[string]int]()
	_, err := p.Parse(`invalid`)

	if err == nil {
		t.Fatal("expected error")
	}
}
