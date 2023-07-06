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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
)

// Option is a functional option for NewTextTemplate. It is used to manipulate
// the TPrompt during Hydrate.
type Option[TPrompt any] func(TPrompt) TPrompt

// NewTextTemplate returns a Prompt that uses a Go Text Template to generate a
// prompt.
func NewTextTemplate[TPrompt, TLLMParams any](
	text string,
	params TLLMParams,
	opts ...Option[TPrompt],
) Prompter[TPrompt, TLLMParams] {
	return &textTemplate[TPrompt, TLLMParams]{
		opts:   opts,
		params: params,
		tmpl: template.Must(template.
			New("prompt").
			Option("missingkey=error").
			Funcs(template.FuncMap{
				"ToJSON": func(v any) string {
					b, err := json.Marshal(v)
					if err != nil {
						panic(err)
					}
					return string(b)
				},
			}).
			Parse(text)),
	}
}

type textTemplate[TPrompt, TLLMParams any] struct {
	tmpl   *template.Template
	params TLLMParams
	opts   []Option[TPrompt]
}

// Hydrate implements Prompter.
func (p *textTemplate[TPrompt, TLLMParams]) Hydrate(ctx context.Context, obj TPrompt) (string, TLLMParams, error) {
	for _, opt := range p.opts {
		obj = opt(obj)
	}

	var buf bytes.Buffer
	if err := p.tmpl.Execute(&buf, obj); err != nil {
		var empty TLLMParams
		return "", empty, fmt.Errorf("%w: %v", ErrHydrate, err)
	}
	return buf.String(), p.params, nil
}
