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

package chains

import (
	"context"
	"fmt"
	"reflect"
)

type Chain[TIn, TOut any] interface {
	Run(context.Context, TIn) (TOut, error)
}

type ChainFunc[TIn, TOut any] func(context.Context, TIn) (TOut, error)

func (c ChainFunc[TIn, TOut]) Run(ctx context.Context, in TIn) (TOut, error) {
	return c(ctx, in)
}

func New[TIn, TOut any](cs ...any) Chain[TIn, TOut] {
	c := chain[TIn, TOut]{
		cs: cs,
	}
	c.validate()
	return c
}

type chain[TIn, TOut any] struct {
	cs []any
}

func (c chain[TIn, TOut]) Run(ctx context.Context, in TIn) (TOut, error) {
	var lastIn any = in

	for _, chain := range c.cs {
		chainValue := reflect.ValueOf(chain)
		methodValue := chainValue.MethodByName("Run")

		// Call the Run method on the chain with the lastIn value as input.
		args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(lastIn)}
		results := methodValue.Call(args)

		// Check if there was an error in the chain execution.
		if errValue := results[1]; !errValue.IsNil() {
			var empty TOut
			return empty, errValue.Interface().(error)
		}

		// Set the lastIn value to the output of the chain.
		lastIn = results[0].Interface()
	}

	// Return the final output value.
	return lastIn.(TOut), nil
}

func (c chain[TIn, TOut]) validate() {
	if len(c.cs) == 0 {
		panic("no chains provided")
	}

	var emptyIn TIn
	var emptyOut TOut
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	var lastOut = reflect.TypeOf(emptyIn)

	for i, chain := range c.cs {
		chainValue := reflect.ValueOf(chain)
		methodValue := chainValue.MethodByName("Run")
		if methodValue.Type().NumIn() != 2 {
			panic(fmt.Sprintf("chain %d: invalid number of arguments", i))
		}
		if actual, expected := methodValue.Type().In(0), ctxType; actual != expected {
			panic(fmt.Sprintf("chain %d: expected input type %v, got %v", i, expected, actual))
		}
		if actual, expected := methodValue.Type().In(1), lastOut; actual != expected {
			panic(fmt.Sprintf("chain %d: expected input type %v, got %v", i, expected, actual))
		}
		if methodValue.Type().NumOut() != 2 {
			panic(fmt.Sprintf("chain %d: invalid number of outputs", i))
		}
		if actual, expected := methodValue.Type().Out(1), errorType; actual != expected {
			panic(fmt.Sprintf("chain %d: invalid output type. Expected %v, got %v", i, expected, actual))
		}

		lastOut = methodValue.Type().Out(0)
	}

	if actual, expected := lastOut, reflect.TypeOf(emptyOut); actual != expected {
		panic(fmt.Sprintf("invalid output type. Expected %v, got %v", expected, actual))
	}
}
