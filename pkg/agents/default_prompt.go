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
	"reflect"

	"github.com/google/go-react/pkg/parsers"
	"github.com/google/go-react/pkg/prompters"
)

const (
	defaultPreamble = "What's the next thing you should do to answer the question with the given tools: "

	defaultPrompt = `{{.Preamble}}

Tools [{{range .Tools}}{{.Name}} {{end}}]:
  {{range .Tools}}{{.Name}}: {{.Description}}{{if gt (len .Args) 0}}

    Usage: {{range .Args}}[{{.}}]{{end}}

    Examples:
    {{range .Examples}}{{.}}
    {{end}}{{end}}
  {{end}}

Rules:
{{range .Rules}} * {{.}}
{{end}}

Format explanation:

  Question: the input question you must answer
  {"thought": "you should always think about what to do and describe your thought process", "action": "the action to take, should be one of [{{range .Tools}}.Name {{end}}]", "input": "the input to the action, it must be included and be on one line.", "observation": "the result of the action. You never add this."}

  ... (this Thought/Action/Action Input/Observation can repeat N times but only add a single iteration)


  {"thought": "I now know the final answer", "final_answer": "the final answer to the original input question. This can only occur if there are no more actions."}

Examples:

  {{range $index, $element := .Examples}}Example {{$index}}:
  Question: {{$element.Question}}
  Previous Context: {{if not $element.PreviousContext}}null{{else}}
  {{range $element.PreviousContext}}{{ToJSON .}}{{end}}{{end}}

  Output:
  {{ToJSON $element.Output}}

  {{end}}

Begin!

Question: {{.Goal}}

Previous context:
{{range .Chains}}{{ToJSON .}}
{{end}}
Output:
`
)

func withDefaultExamples[TLLMParams, TOut any]() prompters.Option[PromptData[TOut]] {
	examples := []PromptDataExample[TOut]{
		{
			Question:        "Add a table",
			PreviousContext: nil,
			Output: Reasoning[TOut]{
				Thought: "I should figure out what the table should be called",
				Action:  "user-input",
				Input:   "What should the table be named?",
			},
		},
	}

	// Check to see if the type is a string, if it is, use the
	// DefaultStringExamples. Otherwise use the DefaultExamples.
	rexamples := reflect.ValueOf(examples)
	var out TOut
	if kind := reflect.ValueOf(&out).Elem().Type().Kind(); kind == reflect.String {
		// We have to use reflection to append these examples because the compiler
		// will try to enforce the type of the slice.
		stringExamples := reflect.ValueOf([]PromptDataExample[string]{
			{
				Question: "Add a table",
				PreviousContext: []ThoughtIteration[string]{
					{
						Reasoning: Reasoning[string]{
							Thought: "I should figure out what the table should be called",
							Action:  "user-input",
							Input:   "What should the table be named?",
						},
						Observation: "employees",
					},
					{
						Reasoning: Reasoning[string]{
							Thought: "I need to add the table employees",
							Action:  "add-table",
							Input:   "employees",
						},
						Observation: "table employees added",
					},
				},
				Output: Reasoning[string]{
					Thought:     "I have finished adding the table employees",
					FinalAnswer: "I have finished adding the table employees",
				},
			},
			{
				Question: "Build a spaceship",
				Output: Reasoning[string]{
					Thought:     "I don't have the tools to build a spaceship",
					FinalAnswer: "I don't have the tools to build a spaceship",
				},
			},
		})
		rexamples = reflect.AppendSlice(rexamples, stringExamples)
		examples = rexamples.Interface().([]PromptDataExample[TOut])
	}

	return WithExamples[TLLMParams, TOut](examples...)
}

var (
	defaultRules = []string{
		"When using a tool, make sure you read the description to ensure it's the right tool and it's used correctly.",
		"If the user asks what you are capable of doing, give them a summary of the tools you have available and what they do.",
		"If the user asks you to do something, make sure you have a tool that can do it. If not, tell the user you can't do it.",
		"When using these tools, if it returns an \"ERROR:\", then the tool failed and needs to be used differently.",
		"Each thought must follow a plan and should be based on previous thoughts and actions.",
		"Use the following JSONL format by only appending a single (thought plus action and input) OR (a thought plus a final answer).",
	}
)

// NewDefaultPrompt returns the default prompt for a ReAct loop. It's highly
// recommened to use your own examples to help the LLM return the proper
// typing and to better capture your use case.
func NewDefaultPrompt[TLLMParams, TOut any](
	params TLLMParams,
	opts ...prompters.Option[PromptData[TOut]],
) prompters.Prompter[PromptData[TOut], TLLMParams] {
	// Set the default examples. We want the user's options to come last so that
	// they can override anything we've set with an option.
	var options []prompters.Option[PromptData[TOut]]
	options = append(options, WithPreamble[TLLMParams, TOut](defaultPreamble))
	options = append(options, WithRules[TLLMParams, TOut](defaultRules...))
	options = append(options, withDefaultExamples[TLLMParams, TOut]())
	options = append(options, withDefaultExamples[TLLMParams, TOut]())

	options = append(options, opts...)

	return prompters.NewTextTemplate[PromptData[TOut], TLLMParams](defaultPrompt, params, options...)
}

// WithPreamble replaces the preamble.
func WithPreamble[TLLMParams, TOut any](preamble string) prompters.Option[PromptData[TOut]] {
	return func(p PromptData[TOut]) PromptData[TOut] {
		p.Preamble = preamble
		return p
	}
}

// WithExamples replaces the default examples with the given examples.
func WithExamples[TLLMParams, TOut any](examples ...PromptDataExample[TOut]) prompters.Option[PromptData[TOut]] {
	return func(p PromptData[TOut]) PromptData[TOut] {
		p.Examples = examples
		return p
	}
}

// WithRules replaces the default rules with the given rules.
func WithRules[TLLMParams, TOut any](rules ...string) prompters.Option[PromptData[TOut]] {
	return func(p PromptData[TOut]) PromptData[TOut] {
		p.Rules = rules
		return p
	}
}

// NewDefaultParser returns the default parser for a ReAct loop.
func NewDefaultParser[TOut any]() parsers.Parser[Reasoning[TOut]] {
	return parsers.NewJSONParser[Reasoning[TOut]]()
}
