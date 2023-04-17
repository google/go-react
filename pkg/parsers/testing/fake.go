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

// Package testing provides a fake parser for testing.
package testing

// Fake is a fake parser for testing.
type Fake[TResp any] struct {
	Datas  []string
	ParseF func(input string) (TResp, error)
}

// Parse implements Parser interface.
func (f *Fake[TResp]) Parse(data string) (TResp, error) {
	f.Datas = append(f.Datas, data)
	if f.ParseF == nil {
		var empty TResp
		return empty, nil
	}
	return f.ParseF(data)
}
