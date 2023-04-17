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

package testing

import (
	"context"

	"github.com/google/go-react/pkg/llms"
)

// Fake implements the llms.LLMS interface for testing.
type Fake[TParams any] struct {
	Prompts    []string
	Params     []TParams
	Outputs    map[string]string
	Err        error
	Errs       map[string]error
	GenerateF  func(ctx context.Context, prompt string)
	AlwaysText string
}

var _ llms.LLM[int] = (*Fake[int])(nil)

// Generate implements the llms.LLMS interface.
func (f *Fake[TParams]) Generate(ctx context.Context, prompt string, params TParams) (string, error) {
	f.Prompts = append(f.Prompts, prompt)
	f.Params = append(f.Params, params)
	if f.Err != nil {
		return "", f.Err
	}
	if f.AlwaysText != "" {
		return f.AlwaysText, nil
	}

	if f.GenerateF != nil {
		f.GenerateF(ctx, prompt)
	}

	return f.Outputs[prompt], f.Errs[prompt]
}
