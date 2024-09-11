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

// Package main defines a CLI for trying out the ReAct concept.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-react/pkg/agents"
	"github.com/google/go-react/pkg/chains/confirmation"
	"github.com/google/go-react/pkg/deepinfra"
	"github.com/google/go-react/pkg/llms"
	"github.com/google/go-react/pkg/llms/vertex"
	"github.com/google/go-react/pkg/predictors"
	"github.com/google/go-react/pkg/prompters"
	"github.com/google/go-react/pkg/tools"
)

var model = flag.String("model", "meta-llama/Meta-Llama-3.1-70B-Instruct", "The model to use for the prompt")

// var apiEndpoint = flag.String("api-endpoint", "us-central1-aiplatform.googleapis.com", "The API endpoint to use")
// var projectID = flag.String("project-id", os.Getenv("GCP_PROJECT_ID"), "The project ID to use")
var maxTokens = flag.Int("max-tokens", 1024, "The maximum number of tokens to generate")
var temperature = flag.Float64("temperature", 0.2, "The temperature to use for the prompt")
var topK = flag.Int("top-k", 40, "The top-k value to use for the prompt")
var topP = flag.Float64("top-p", 0.9, "The top-p value to use for the prompt")

func main() {
	flag.Parse()
	ctx := context.Background()

	defaultParams := vertex.Params{
		Model:       *model,
		MaxTokens:   *maxTokens,
		Temperature: *temperature,
		TopK:        *topK,
		TopP:        *topP,
	}
	prompt := agents.NewDefaultPrompt[vertex.Params, string](defaultParams)
	tempFile, err := os.CreateTemp("", "react")
	if err != nil {
		log.Fatalf("failed to create temp file: %s", err)
	}
	defer os.Remove(tempFile.Name())
	log.Printf("Using temp file for prompts: %s", tempFile.Name())
	prompt = prompters.NewLogger(prompt, tempFile)

	llm := getLLMDeepInfra(ctx)

	parser := agents.NewDefaultParser[string]()
	predictor := predictors.New(llm, prompt, parser)
	predictor = agents.NewCLILogger(predictor, os.Stderr)
	tools := SetupToyToolSet(llm)

	for {
		fmt.Println("AI: What is the goal?")
		fmt.Print("You: ")
		goal, err := ReadLine()
		if err != nil {
			log.Fatalf("failed to read line: %v", err)
		}

		finalAnswer, err := agents.NewAgent(predictor, tools...).Run(ctx, goal)
		if err != nil {
			log.Fatalf("agent.Run failed: %v", err)
		}
		fmt.Println(finalAnswer)
	}

}

func getLLMDeepInfra(_ context.Context) llms.LLM[vertex.Params] {
	infra := deepinfra.NewDeepInfraLLM("pgT0g5nBK0FOHgLuhTo8UxEqulaAaptm")
	return infra
}

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

// Package toytoolset defines a set of tools that are used to demo the react agent.

// SetupToyToolSet returns a tool set of demo tools.
func SetupToyToolSet(llm llms.LLM[vertex.Params]) []tools.Tool {
	var tools []tools.Tool
	var appTemplate appTemplate

	tools = append(tools,
		userInputTool(),
		listTablesTool(&appTemplate),
		addTableTool(llm, &appTemplate),
		removeTableTool(llm, &appTemplate),
		saveAppTemplateTool(&appTemplate),
	)
	return tools
}

type appTemplate struct {
	Tables []appTable
}

type appTable struct {
	Name    string
	Columns []appColumn
}

type appColumn struct {
	Name string
	Type string
}

// ReadLine returns a line from the user.
func ReadLine() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", scanner.Err()
	}
	return "", errors.New("unable to get input from user")
}

func userInputTool() tools.Tool {
	return tools.Tool{
		Name:        "user-input",
		Description: "Ask the user a question. The input is what is displayed to the user.",
		Examples:    []string{"some question to the user"},
		Run: func(ctx context.Context, input string) (any, error) {
			fmt.Printf("AI: %s\n", input)

			// Read a line from the user.
			fmt.Print("You: ")
			return ReadLine()
		},
	}
}

func listTablesTool(appTemplate *appTemplate) tools.Tool {
	return tools.Tool{
		Name:        "list-tables",
		Description: "List the tables in the app. It does not take any input.",
		Run: func(ctx context.Context, input string) (any, error) {
			var names []string
			for _, table := range appTemplate.Tables {
				names = append(names, table.Name)
			}
			return strings.Join(names, ","), nil
		},
	}
}

func addTableTool(llm llms.LLM[vertex.Params], appTemplate *appTemplate) tools.Tool {
	confirmationChain := confirmation.New(llm, os.Stdin, os.Stdout)
	return tools.Tool{
		Name:        "add-table",
		Description: "Add a table within the app. The input is the name of the table.",
		Examples: []string{
			"some-table-name",
		},
		Args: []string{
			"table name",
		},
		Run: func(ctx context.Context, input string) (any, error) {
			parts := strings.Fields(input)
			if len(parts) != 1 {
				return "", fmt.Errorf("%w: should only get one input", tools.ErrInvalidToolInput)
			}
			tableName := parts[0]

			confirm, err := confirmationChain.Run(ctx, fmt.Sprintf("Adding table %s", tableName))
			if err != nil {
				return "", err
			} else if confirm.Error != "" {
				return "", errors.New(confirm.Error)
			} else if !confirm.Result {
				return "", errors.New("user changed their mind")
			}

			appTemplate.Tables = append(appTemplate.Tables, appTable{Name: tableName})
			return fmt.Sprintf("Table %s added", tableName), nil
		},
	}
}

func removeTableTool(llm llms.LLM[vertex.Params], appTemplate *appTemplate) tools.Tool {
	confirmationChain := confirmation.New(llm, os.Stdin, os.Stdout)
	return tools.Tool{
		Name:        "remove-table",
		Description: "Remove a table within the app. The input is the name of the table.",
		Examples: []string{
			"some-table-name",
		},
		Args: []string{
			"table name",
		},
		Run: func(ctx context.Context, input string) (any, error) {
			parts := strings.Fields(input)
			if len(parts) != 1 {
				return "", fmt.Errorf("%w: should only get one input", tools.ErrInvalidToolInput)
			}
			tableName := parts[0]

			confirm, err := confirmationChain.Run(ctx, fmt.Sprintf("Removing table %s", tableName))
			if err != nil {
				return "", err
			} else if confirm.Error != "" {
				return "", errors.New(confirm.Error)
			} else if !confirm.Result {
				return "", errors.New("user changed their mind")
			}

			for i, table := range appTemplate.Tables {
				if table.Name != tableName {
					continue
				}
				appTemplate.Tables = append(appTemplate.Tables[:i], appTemplate.Tables[i+1:]...)
				return fmt.Sprintf("Table %s removed", tableName), nil
			}

			return "failed", fmt.Errorf("table %s not found", tableName)
		},
	}
}

func saveAppTemplateTool(appTemplate *appTemplate) tools.Tool {
	return tools.Tool{
		Name:        "save-app-template",
		Description: "Save the app template. The input is the file name.",
		Examples: []string{
			"some/file/name",
		},
		Args: []string{
			"file name",
		},
		Run: func(ctx context.Context, input string) (any, error) {
			parts := strings.Fields(input)
			if len(parts) != 1 {
				return "", fmt.Errorf("%w: should only get one input", tools.ErrInvalidToolInput)
			}
			fileName := parts[0]

			f, err := os.Create(fileName)
			if err != nil {
				return "", err
			}
			defer f.Close()

			if err := json.NewEncoder(f).Encode(appTemplate); err != nil {
				return "", err
			}
			return "successfully wrote AppTemplate", nil
		},
	}
}
