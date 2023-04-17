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
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-react/pkg/agents"
	"github.com/google/go-react/pkg/llms"
	"github.com/google/go-react/pkg/llms/vertex"
	"github.com/google/go-react/pkg/predictors"
	"github.com/google/go-react/pkg/prompters"
)

var model = flag.String("model", "text-bison@001", "The model to use for the prompt")
var apiEndpoint = flag.String("api-endpoint", "us-central1-aiplatform.googleapis.com", "The API endpoint to use")
var projectID = flag.String("project-id", os.Getenv("GCP_PROJECT_ID"), "The project ID to use")
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

	llm := getLLM(ctx)

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

func getLLM(ctx context.Context) llms.LLM[vertex.Params] {
	if *projectID == "" {
		log.Fatalf("you must set the project-id flag or GCP_PROJECT_ID environment variable")
	}

	llm, err := vertex.New(ctx, *apiEndpoint, *projectID)
	if err != nil {
		log.Fatalf("failed to create LLM: %v", err)
	}
	return llm
}
