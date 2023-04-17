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

package predictors

import (
	"context"
	"errors"
)

type retrier[TReq, TResp any] struct {
	p Predictor[TReq, TResp]
}

// NewRetrier returns a Predictor that wraps the given Predictor. It will retry
// on certain types of errors.
func NewRetrier[TReq, TResp any](
	p Predictor[TReq, TResp],
) Predictor[TReq, TResp] {
	// TODO: We should likely support some configuration here at some point.
	return retrier[TReq, TResp]{
		p: p,
	}
}

// Predict implements Predictor.
func (r retrier[TReq, TResp]) Predict(ctx context.Context, req TReq) (TResp, error) {
	var resp TResp
	var err error
	for i := 0; i < 3; i++ {
		resp, err = r.p.Predict(ctx, req)
		if errors.Is(err, ErrLLM) {
			// TODO: We should have metrics around this.
			continue
		}
		if errors.Is(err, ErrParse) {
			// TODO: We should have metrics around this.
			continue
		}

		// Error, if any, is not a retryable one.
		return resp, err
	}

	// Retying failed, return the last error.
	return resp, err
}
