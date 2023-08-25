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

package userinput

import (
	"context"
	"fmt"

	"github.com/google/go-react/pkg/tools"
)

func New() tools.Tool {
	return tools.Tool{
		Name:        "user input",
		Description: "used to ask a question from the user. Do not use to simply print a message to the user.",
		Args:        []string{"prompt"},
		Run: func(ctx context.Context, input string) (any, error) {
			fmt.Printf("AI: %s\n", input)

			// Read a line from the user.
			fmt.Print("You: ")
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				return "", err
			}
			return response, nil
		},
	}
}
