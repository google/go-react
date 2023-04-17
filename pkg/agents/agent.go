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
	"errors"
	"fmt"

	"github.com/google/go-react/pkg/predictors"
	"github.com/google/go-react/pkg/tools"
)

// ErrInvalidTool is returned when a tool is not found.
var ErrInvalidTool = errors.New("invalid tool")

// Agent reasons and acts using its given tools and LLM.
type Agent[TOut any] struct {
	p         predictors.Predictor[PromptData[TOut], Reasoning[TOut]]
	tools     map[string]tools.Tool
	toolSlice []tools.Tool
}

// Reasoning is the expected output from a LLM where it reasons about the given
// problem.
type Reasoning[TOut any] struct {
	Thought     string `json:"thought,omitempty"`
	Action      string `json:"action,omitempty"`
	Input       string `json:"input,omitempty"`
	FinalAnswer TOut   `json:"final_answer,omitempty"`
}

// ThoughtIteration is used to store an Observation along with the Reasoning.
// The Observation comes from the requested Action and its associated Input.
type ThoughtIteration[TOut any] struct {
	Reasoning[TOut] `json:",inline"`
	Observation     string `json:"observation,omitempty"`
}

// PromptData is the data used to hydrate the ReAct prompt.
type PromptData[TOut any] struct {
	Tools  []tools.Tool
	Goal   string
	Chains []ThoughtIteration[TOut]
}

// NewAgent returns a new ReAct Agent. It properly wraps the basePredictor in the
// necessary retry and extra validation logic.
func NewAgent[TOut any](
	basePredictor predictors.Predictor[PromptData[TOut], Reasoning[TOut]],
	toolSet ...tools.Tool,
) Agent[TOut] {
	if len(toolSet) == 0 {
		panic("no tools provided")
	}
	toolsM, toolsS, err := validateTools(toolSet)
	if err != nil {
		panic(err)
	}

	// Chain together the predictors.
	// This chain will first ensure the Reasoning object follows the necessary
	// rules (e.g., has thought) and will then retry on errors.
	p := predictors.NewRetrier(newReasoningPredictor(basePredictor, toolsS))

	return Agent[TOut]{
		p:         p,
		tools:     toolsM,
		toolSlice: toolsS,
	}
}

func validateTools(ts []tools.Tool) (map[string]tools.Tool, []tools.Tool, error) {
	m := map[string]tools.Tool{}
	var toolsSlice []tools.Tool

	for i, t := range ts {
		name := normalizeToolName(t.Name)
		if name == "" {
			return nil, nil, fmt.Errorf("%d: tool name is empty", i)
		}
		if t.Description == "" {
			return nil, nil, fmt.Errorf("%d: tool description is empty", i)
		}
		if _, ok := m[name]; ok {
			return nil, nil, fmt.Errorf("%d: multiple tools with the same name were used: %q", i, name)
		}

		m[name] = t
		t.Name = name
		toolsSlice = append(toolsSlice, t)
	}

	return m, toolsSlice, nil
}

// Run a ReAct Agent in attempt a to reach the given goal. The FinalAnswer is
// output unless there is an error first.
func (a Agent[TOut]) Run(ctx context.Context, goal string) (TOut, error) {
	if goal == "" {
		var empty TOut
		return empty, errors.New("goal is empty")
	}

	// Iterations is the history of the ReAct loop. These are built up over each
	// loop and sent back in the prompt.
	var iterations []ThoughtIteration[TOut]

	for {
		promptData := PromptData[TOut]{
			Goal:   goal,
			Tools:  a.toolSlice,
			Chains: iterations,
		}
		reasoning, err := a.p.Predict(ctx, promptData)
		if errors.Is(err, ErrInvalidTool) {
			// The tool was invalid, so we need to prompt again.
			iterations = append(iterations, ThoughtIteration[TOut]{
				Reasoning: reasoning,
				// Observation: errorObservation(errors.Unwrap(err)),
				Observation: errorObservation(err),
			})
			continue
		} else if err != nil {
			var empty TOut
			return empty, err
		}
		if reasoning.Action == "" {
			// We have the FinalAnswer, return it.
			return reasoning.FinalAnswer, nil
		}

		// We are acting on an Action.

		// NOTE: We know the tool exists because our predictor chain asserts that it does.
		tool := a.tools[normalizeToolName(reasoning.Action)]
		observation, err := tool.Run(ctx, reasoning.Input)
		if err != nil {
			// The tool errored, so we need to prompt again.
			iterations = append(iterations, ThoughtIteration[TOut]{
				Reasoning:   reasoning,
				Observation: errorObservation(err),
			})
			continue
		}

		// Save the result from the tool.
		iterations = append(iterations, ThoughtIteration[TOut]{
			Reasoning:   reasoning,
			Observation: observation,
		})
	}
}

func errorObservation(err error) string {
	return fmt.Sprintf("ERROR: %v", err)
}
