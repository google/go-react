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
	"encoding/json"
	"io"
)

type jsonLogger[TReq, TResp any] struct {
	p       Predictor[TReq, TResp]
	encoder *json.Encoder
}

// NewJSONLogger chains a Predictor that provides logging around the input and
// output of the Predictor.
func NewJSONLogger[TReq, TResp any](
	p Predictor[TReq, TResp],
	out io.Writer,
) Predictor[TReq, TResp] {
	return jsonLogger[TReq, TResp]{
		p:       p,
		encoder: json.NewEncoder(out),
	}
}

// Predict implements Predictor.
func (l jsonLogger[TReq, TResp]) Predict(ctx context.Context, req TReq) (TResp, error) {
	if err := l.encoder.Encode(map[string]any{"request": req}); err != nil {
		var empty TResp
		return empty, err
	}
	resp, err := l.p.Predict(ctx, req)
	if err != nil {
		return resp, err
	}
	if err := l.encoder.Encode(map[string]any{"response": resp}); err != nil {
		var empty TResp
		return empty, err
	}
	return resp, nil
}
