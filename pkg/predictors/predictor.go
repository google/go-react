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

package predictors

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-react/pkg/llms"
	"github.com/google/go-react/pkg/parsers"
	"github.com/google/go-react/pkg/prompters"
)

var (
	// ErrLLM is returned when prompting the LLM fails. This error will be retried
	// on.
	ErrLLM = errors.New("failed to obtain response from LLM")
	// ErrParse is returned when the response from the LLM fails to parse. This
	// error will be retried on.
	ErrParse = errors.New("failed to parse response")
)

// Predictor has a Predict method that will be used to predict responses from the LLM.
// This interface can be chained with other LLM implementations for things like
// retries and/or logging.
type Predictor[TReq, TResp any] interface {
	// Predict submits the given request to the LLM and fetches a response.
	Predict(ctx context.Context, req TReq) (TResp, error)
}

type predictor[TReq, TResp, TLLMParams any] struct {
	model    llms.LLM[TLLMParams]
	prompter prompters.Prompter[TReq, TLLMParams]
	parser   parsers.Parser[TResp]
}

// New returns a Predictor from the given LLM.
func New[TReq, TResp, TLLMParams any](
	model llms.LLM[TLLMParams],
	prompter prompters.Prompter[TReq, TLLMParams],
	parser parsers.Parser[TResp],
) Predictor[TReq, TResp] {
	return predictor[TReq, TResp, TLLMParams]{
		model:    model,
		prompter: prompter,
		parser:   parser,
	}
}

// Predict implements Predictor.
func (p predictor[TReq, TResp, TLLMParams]) Predict(ctx context.Context, req TReq) (TResp, error) {
	var empty TResp
	prompt, params, err := p.prompter.Hydrate(ctx, req)
	if err != nil {
		return empty, fmt.Errorf("%w: %v", prompters.ErrHydrate, err)
	}

	llmOutput, err := p.model.Generate(ctx, prompt, params)
	if err != nil {
		return empty, fmt.Errorf("%w: %v", ErrLLM, err)
	}

	result, err := p.parser.Parse(llmOutput)
	if err != nil {
		return empty, fmt.Errorf("%w: %v", ErrParse, err)
	}

	return result, nil
}
