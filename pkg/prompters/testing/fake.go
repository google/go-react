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

// Package testing contains the fake prompter for testing.
package testing

import "context"

// Fake is a fake prompter for testing.
type Fake[TPrompt, TLLMParams any] struct {
	HydrateData []TPrompt
	HydrateF    func(ctx context.Context, vars TPrompt) (string, TLLMParams, error)
}

// Hydrate hydrates the fake prompter.
func (f *Fake[TPrompt, TLLMParams]) Hydrate(ctx context.Context, data TPrompt) (string, TLLMParams, error) {
	f.HydrateData = append(f.HydrateData, data)
	if f.HydrateF == nil {
		var empty TLLMParams
		return "", empty, nil
	}
	return f.HydrateF(ctx, data)
}
