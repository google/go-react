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

package tools

import (
	"context"
	"errors"
)

// ErrInvalidToolInput is returned when the tool's input requirements are not
// satisfied.
var ErrInvalidToolInput = errors.New("invalid tool input")

// Tool is a struct that describes the given tool.
type Tool struct {
	Name        string                                                  `json:"Name,omitempty"`
	Description string                                                  `json:"Description,omitempty"`
	Examples    []string                                                `json:"Examples,omitempty"`
	Args        []string                                                `json:"Args,omitempty"`
	Run         func(ctx context.Context, input string) (string, error) `json:"-"`
}
