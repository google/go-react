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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-react/pkg/llms"
	"github.com/google/go-react/pkg/llms/vertex"
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
	log.SetFlags(0)

	if len(flag.Args()) != 1 {
		log.Fatal("usage: prompt <text>")
	}

	prompt := flag.Arg(0)

	// Check to see if the prompt is actually a file.
	if strings.HasPrefix(prompt, "@") {
		b, err := ioutil.ReadFile(prompt[1:])
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		prompt = string(b)
	}

	defaultParams := vertex.Params{
		Model:       *model,
		MaxTokens:   *maxTokens,
		Temperature: *temperature,
		TopK:        *topK,
		TopP:        *topP,
	}
	llm := getLLM(context.Background())
	resp, err := llm.Generate(context.Background(), prompt, defaultParams)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
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
