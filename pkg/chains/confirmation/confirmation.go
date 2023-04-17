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

package confirmation

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/google/go-react/pkg/chains"
	"github.com/google/go-react/pkg/llms"
	"github.com/google/go-react/pkg/parsers"
	"github.com/google/go-react/pkg/prompters"
)

// Response is the response from the user.
type Response struct {
	Result bool   `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// New returns a chains.Chain that will prompt the user for confirmation.
func New[TLLMParams any](
	llm llms.LLM[TLLMParams],
	in io.Reader,
	out io.Writer,
) chains.Chain[string, Response] {
	return chains.ChainFunc[string, Response](func(ctx context.Context, input string) (Response, error) {
		fmt.Fprintf(out, confirmationUserPromptTemplate, input)

		reader := bufio.NewReader(in)
		text, err := reader.ReadString('\n')
		if err != nil {
			return Response{}, fmt.Errorf("failed to read confirmation: %v", err)
		}

		prompter := newPrompter[TLLMParams]()
		parser := newParser()
		p, params, err := prompter.Hydrate(ctx, request{Input: text})
		if err != nil {
			return Response{}, fmt.Errorf("failed to hydrate confirmation prompt: %v", err)
		}
		resp, err := llm.Generate(ctx, p, params)

		result, err := parser.Parse(resp)
		if err != nil {
			return Response{}, fmt.Errorf("failed to parse confirmation prompt: %v", err)
		}

		return result, nil
	})
}

const (
	confirmationUserPromptTemplate = "%s\n\nDo you want to continue?: "

	confirmationPrompt = `Check if the input signals that the user said some form of yes.

Example 1:
Input: "yes"
Output: {"result": true}

Example 2:
Input: "no"
Output: {"result": false}

Example 3:
Input: "yea"
Output: {"result": true}

Example 4:
Input: "nevermind, lets add a table instead"
Output: {"error": "user wants to do something different: lets add a table instead"}

Begin!

Input: {{.Input}}
Output:
`
)

type request struct {
	Input string
}

func newPrompter[TLLMParams any]() prompters.Prompter[request, TLLMParams] {
	var defaultParams TLLMParams
	return prompters.NewTextTemplate[request, TLLMParams](confirmationPrompt, defaultParams)
}

func newParser() parsers.Parser[Response] {
	return parsers.NewJSONParser[Response]()
}
