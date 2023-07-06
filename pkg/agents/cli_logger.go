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
	"io"

	"github.com/google/go-react/pkg/predictors"
)

// CLILoggerOption is an option for NewCLILogger.
type CLILoggerOption[TOut any] func(*cliLogger[TOut])

func WithCLILoggerPrefix[TOut any](prefix string) func(*cliLogger[TOut]) {
	return func(c *cliLogger[TOut]) {
		c.prefix = prefix
	}
}

type cliLogger[TOut any] struct {
	p      predictors.Predictor[PromptData[TOut], Reasoning[TOut]]
	out    io.Writer
	prefix string
}

// NewCLILogger chains a Predictor that provides logging around the output of
// the Predictor that is suitable for a CLI.
func NewCLILogger[TOut any](
	p predictors.Predictor[PromptData[TOut], Reasoning[TOut]],
	out io.Writer,
	opts ...CLILoggerOption[TOut],
) predictors.Predictor[PromptData[TOut], Reasoning[TOut]] {
	l := cliLogger[TOut]{
		p:   p,
		out: out,
	}
	for _, opt := range opts {
		opt(&l)
	}
	return l
}

// Predict implements Predictor.
func (l cliLogger[TOut]) Predict(
	ctx context.Context,
	req PromptData[TOut],
) (Reasoning[TOut], error) {
	resp, err := l.p.Predict(ctx, req)
	if err != nil {
		return resp, err
	}

	const (
		// ANSI escape sequence for setting text color to green.
		green = "\033[32m"

		// ANSI escape sequence for setting text color to blue.
		blue = "\033[34m"

		// ANSI escape sequence for resetting text color to default.
		reset = "\033[0m"
	)

	prefix := ""
	if l.prefix != "" {
		// Print the prefix in blue.
		prefix = fmt.Sprintf("%s[%s]%s ", blue, l.prefix, reset)
	}

	// Print out in green.
	if resp.Action != "" {
		fmt.Fprintf(l.out, "%s%sThought: %q Action: %q Input: %q%s\n",
			prefix, green, resp.Thought, resp.Action, resp.Input, reset)
	} else {
		fmt.Fprintf(l.out, "%s%sThought: %q FinalAnswer: %v%s\n",
			prefix, green, resp.Thought, resp.FinalAnswer, reset)
	}
	return resp, nil
}
