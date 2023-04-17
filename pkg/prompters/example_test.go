package prompters_test

import (
	"context"
	"fmt"

	"github.com/google/go-react/pkg/prompters"
)

func ExampleTextTemplate() {
	type Data struct {
		Product string
	}
	data := Data{Product: "Gophers"}
	llmParams := 3

	tmpl := "Come up with store names that sell {{.Product}}!"
	t := prompters.NewTextTemplate[Data, int](tmpl, llmParams)

	prompt, params, err := t.Hydrate(context.Background(), data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Prompt: %s\nParams: %d\n", prompt, params)

	// Output:
	// Prompt: Come up with store names that sell Gophers!
	// Params: 3
}
