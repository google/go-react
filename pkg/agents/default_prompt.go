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
	"github.com/google/go-react/pkg/parsers"
	"github.com/google/go-react/pkg/prompters"
)

const (
	defaultPrompt = `What's the next thing you should do to answer the question with the given tools.

Tools:
  {{range .Tools}}{{.Name}}: {{.Description}}{{if gt (len .Args) 0}}

    Usage: {{range .Args}}[{{.}}]{{end}}

    Examples:
    {{range .Examples}}{{.}}
    {{end}}{{end}}
  {{end}}

Rules:
* When using a tool, make sure you read the description to ensure it's the right tool and it's used correctly.
* If the user asks what you are capable of doing, give them a summary of the tools you have available and what they do.
* If the user asks you to do something, make sure you have a tool that can do it. If not, tell the user you can't do it.
* When using these tools, if it returns an "ERROR:", then the tool failed and needs to be used differently.
* Each thought must follow a plan and should be based on previous thoughts and actions.
* If you get stuck, ask the user a question via the user-input tool.
* Use the following JSONL format by only appending a single (thought plus action and input) OR (a thought plus a final answer).

Format explanation:

  Question: the input question you must answer
  {"thought": "you should always think about what to do and describe your thought process", "action": "the action to take, should be one of [{{range .Tools}}.Name {{end}}]", "input": "the input to the action, it must be included and be on one line.", "observation": "the result of the action. You never add this."}

  ... (this Thought/Action/Action Input/Observation can repeat N times but only add a single iteration)


  {"thought": "I now know the final answer", "final_answer": "the final answer to the original input question. This can only occur if there are no more actions."}

Examples:

  Example 1:
  Input:
  Question: Add a table

	Previous context: null

  Output:
  {"thought": "I should figure out what the table should be called", "action": "user-input", "input": "What should the table be named?", "observation": ""}

  Example 2:
  Input:
  Question: Help the user edit their app

	Previous context:
  {"thought": "I should figure out what the table should be called", "action": "user-input", "input": "What should the table be named?", "observation": "employees"}
  {"thought": "I need to add the table employees", "action": "add-table", "input": "employees", "observation": ""}

  Output:
  {"thought": "I have finished adding the table employees", "final_answer": "I have finished adding the table employees"}

  Example 3:
  Input:
  Question: Build me a space ship

	Previous context: null

  Output:
  {"thought": "I don't have the tools to build a space ship", "final_answer": "I don't have the tools to build a space ship"}


Begin!

Input:
Question: {{.Goal}}

Previous context:
{{range .Chains}}{{ToJSON .}}
{{end}}
Output:
`
)

// NewDefaultPrompt returns the default prompt for a ReAct loop.
func NewDefaultPrompt[TLLMParams, TOut any](params TLLMParams) prompters.Prompter[PromptData[TOut], TLLMParams] {
	return prompters.NewTextTemplate[PromptData[TOut], TLLMParams](defaultPrompt, params)
}

// NewDefaultParser returns the default parser for a ReAct loop.
func NewDefaultParser[TOut any]() parsers.Parser[Reasoning[TOut]] {
	return parsers.NewJSONParser[Reasoning[TOut]]()
}
