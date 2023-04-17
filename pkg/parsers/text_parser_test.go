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

func TestTextParser(t *testing.T) {
	t.Parallel()
	p := parsers.NewTextParser()
	val, err := p.Parse("some text")
	if err != nil {
		t.Fatal(err)
	}
	if actual, expected := val, "some text"; actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
