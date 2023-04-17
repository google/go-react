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

package prompters

import (
	"context"
	"errors"
)

var (
	// ErrHydrate is returned when hydrating the prompt fails. This error will not
	// be retried on.
	ErrHydrate = errors.New("failed to hydrate prompt")
)

// Prompter is an interface for a generic prompt that can be hydrated with the
// given type.
type Prompter[TPrompt, TLLMParams any] interface {
	// Hydrate hydrates the prompt with the given type.
	Hydrate(context.Context, TPrompt) (string, TLLMParams, error)
}
