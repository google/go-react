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

package parsers

import (
	"encoding/json"
	"strings"
)

type jsonParser[T any] struct {
}

// NewJSONParser returns a Parser that parses JSON to the given type.
func NewJSONParser[T any]() Parser[T] {
	return &jsonParser[T]{}
}

// Parse implements Parser.
func (p *jsonParser[T]) Parse(data string) (T, error) {
	var result T
	err := json.NewDecoder(strings.NewReader(data)).Decode(&result)
	return result, err
}
