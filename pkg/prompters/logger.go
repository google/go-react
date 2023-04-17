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
	"fmt"
	"io"
)

type logger[TPrompt, TLLMParams any] struct {
	p   Prompter[TPrompt, TLLMParams]
	out io.Writer
}

// NewLogger writes each hydrated prompt to the given io.Writer.
func NewLogger[TPrompt, TLLMParams any](p Prompter[TPrompt, TLLMParams], out io.Writer) Prompter[TPrompt, TLLMParams] {
	return logger[TPrompt, TLLMParams]{
		p:   p,
		out: out,
	}
}

// Hydrate implements Prompter.
func (l logger[TPrompt, TLLMParams]) Hydrate(ctx context.Context, p TPrompt) (string, TLLMParams, error) {
	output, params, err := l.p.Hydrate(ctx, p)
	fmt.Fprintln(l.out, output)
	return output, params, err
}
