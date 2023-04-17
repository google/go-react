package predictors_test

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-react/pkg/predictors"
	predictorstesting "github.com/google/go-react/pkg/predictors/testing"
)

func ExampleChain() {
	// This example demonstrates how to chain multiple predictors together.
	// The fake one will always return an error indicating that the prediction
	// from the LLM failed for some reason. The JSONLogger will log the request
	// while the retrier will retry the request 3 times before giving up.

	var predictor predictors.Predictor[int, string] = &predictorstesting.Fake[int, string]{
		Err: fmt.Errorf("%w: some-error", predictors.ErrLLM),
	}
	predictor = predictors.NewJSONLogger(predictor, os.Stdout)
	predictor = predictors.NewRetrier(predictor)

	predictor.Predict(context.Background(), 1)

	// Output:
	// {"request":1}
	// {"request":1}
	// {"request":1}
}
