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

package chains_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-react/pkg/chains"
)

func TestNew_single(t *testing.T) {
	t.Parallel()

	c := chains.New[int, int](addOne)
	result, err := c.Run(context.Background(), 1)

	if err != nil {
		t.Error(err)
	}

	if actual, expected := result, 2; actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}
}

func TestNew_multi(t *testing.T) {
	t.Parallel()

	c := chains.New[int, string](addOne, concatHello)
	result, err := c.Run(context.Background(), 1)

	if err != nil {
		t.Error(err)
	}

	if actual, expected := result, "2 hello"; actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestNew_bails_on_error(t *testing.T) {
	t.Parallel()

	c := chains.New[int, int](alwaysErrors, alwaysPanics)
	_, err := c.Run(context.Background(), 1)

	if err == nil {
		t.Error("expected error")
	}
}

func TestNew_empty(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, int]()
}

func TestNew_single_invalid_input(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[string, int](addOne)
}

func TestNew_single_invalid_output(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, string](addOne)
}

func TestNew_multi_invalid_input(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, string](concatGoodbye, concatGoodbye)
}

func TestNew_multi_invalid_output(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, string](addOne, addOne)
}

func TestNew_invalid_run_extra_in(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, int](wrongNumInArgs{})
}

func TestNew_invalid_run_wrong_ctx(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, int](wrongCtxArg{})
}

func TestNew_invalid_run_extra_out(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, int](wrongNumOutArgs{})
}

func TestNew_invalid_run_wrong_err(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, int](wrongErrArgs{})
}

func TestNew_multi_invalid_middle(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	chains.New[int, int](addOne, concatHello, addOne)
}

var addOne = chains.ChainFunc[int, int](func(ctx context.Context, in int) (int, error) {
	return in + 1, nil
})

var concatHello = chains.ChainFunc[int, string](func(ctx context.Context, in int) (string, error) {
	return fmt.Sprintf("%d hello", in), nil
})

var concatGoodbye = chains.ChainFunc[string, string](func(ctx context.Context, in string) (string, error) {
	return fmt.Sprintf("%s hello", in), nil
})

var alwaysErrors = chains.ChainFunc[int, int](func(ctx context.Context, in int) (int, error) {
	return 0, fmt.Errorf("always errors")
})

var alwaysPanics = chains.ChainFunc[int, int](func(ctx context.Context, in int) (int, error) {
	panic("always panics")
})

type wrongNumInArgs struct{}

func (w wrongNumInArgs) Run(ctx context.Context, in, extra int) (int, error) {
	panic("not implemented")
}

type wrongCtxArg struct{}

func (w wrongCtxArg) Run(in, extra int) (int, error) {
	panic("not implemented")
}

type wrongNumOutArgs struct{}

func (w wrongNumOutArgs) Run(ctx context.Context, in int) (int, int, error) {
	panic("not implemented")
}

type wrongErrArgs struct{}

func (w wrongErrArgs) Run(ctx context.Context, in int) (int, int) {
	panic("not implemented")
}
