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

package agents

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/go-react/pkg/predictors"
	"github.com/google/go-react/pkg/tools"
)

var (
	toolSpacePattern = regexp.MustCompile(`[_\s]+`)
)

type reasoningPredictor[TOut any] struct {
	p     predictors.Predictor[PromptData[TOut], Reasoning[TOut]]
	tools map[string]string
}

func newReasoningPredictor[TOut any](
	p predictors.Predictor[PromptData[TOut], Reasoning[TOut]],
	tools []tools.Tool,
) predictors.Predictor[PromptData[TOut], Reasoning[TOut]] {

	ts := map[string]string{}
	for _, t := range tools {
		name := normalizeToolName(t.Name)
		ts[name] = t.Name
	}

	return reasoningPredictor[TOut]{
		p:     p,
		tools: ts,
	}
}

// Predict implements Predictor.
func (p reasoningPredictor[TOut]) Predict(
	ctx context.Context,
	req PromptData[TOut],
) (Reasoning[TOut], error) {
	r, err := p.p.Predict(ctx, req)
	if err != nil {
		return Reasoning[TOut]{}, err
	}

	// Assert that we have a Thought and either an Action or a FinalAnswer.
	if r.Thought == "" {
		return Reasoning[TOut]{}, fmt.Errorf("%w: Missing Thought field", predictors.ErrParse)
	}

	defaultValue := reflect.Zero(reflect.TypeOf(r.FinalAnswer)).Interface()
	isDefaultFinalAnswer := reflect.DeepEqual(r.FinalAnswer, defaultValue)

	if r.Action == "" && isDefaultFinalAnswer {
		return Reasoning[TOut]{}, fmt.Errorf("%w: Either Action or FinalAnswer must be set", predictors.ErrParse)
	}
	if r.Action != "" && !isDefaultFinalAnswer {
		return Reasoning[TOut]{}, fmt.Errorf("%w: Both Action and FinalAnswer are set", predictors.ErrParse)
	}

	if r.Action != "" {
		// Replace the Action with the tool name to ensure casing is correct.
		r.Action = normalizeToolName(r.Action)

		// Assert that we know what tool it is referencing.
		if _, ok := p.tools[r.Action]; !ok {
			return Reasoning[TOut]{}, fmt.Errorf("%w: %q", ErrInvalidTool, r.Action)
		}
	}

	return r, err
}

func normalizeToolName(name string) string {
	name = strings.ToLower(name)
	name = toolSpacePattern.ReplaceAllString(name, "-")
	return name
}
