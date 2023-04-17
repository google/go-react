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

// Package testing contains testing utilities for the predictors package.
package testing

import "context"

// Fake is a fake Predictor that can be used in tests.
type Fake[TReq, TResp any] struct {
	Reqs  []TReq
	Resps []TResp
	Err   error
}

// Predict implements Predictor.
func (f *Fake[TReq, TResp]) Predict(ctx context.Context, req TReq) (TResp, error) {
	f.Reqs = append(f.Reqs, req)

	if f.Err != nil {
		var empty TResp
		return empty, f.Err
	}

	resp := f.Resps[0]
	f.Resps = f.Resps[1:]
	return resp, nil
}
