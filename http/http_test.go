// Copyright 2014 The Cayley Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/cayley/quad"
)

var parseTests = []struct {
	message string
	input   string
	expect  []*quad.Quad
	err     error
}{
	{
		message: "parse correct JSON",
		input: `[
			{"subject": "foo", "predicate": "bar", "object": "baz"},
			{"subject": "foo", "predicate": "bar", "object": "baz", "label": "graph"}
		]`,
		expect: []*quad.Quad{
			{"foo", "bar", "baz", ""},
			{"foo", "bar", "baz", "graph"},
		},
		err: nil,
	},
	{
		message: "parse correct JSON with extra field",
		input: `[
			{"subject": "foo", "predicate": "bar", "object": "foo", "something_else": "extra data"}
		]`,
		expect: []*quad.Quad{
			{"foo", "bar", "foo", ""},
		},
		err: nil,
	},
	{
		message: "reject incorrect JSON",
		input: `[
			{"subject": "foo", "predicate": "bar"}
		]`,
		expect: nil,
		err:    fmt.Errorf("Invalid triple at index %d. %v", 0, &quad.Quad{"foo", "bar", "", ""}),
	},
}

func TestParseJSON(t *testing.T) {
	for _, test := range parseTests {
		got, err := ParseJsonToTripleList([]byte(test.input))
		if fmt.Sprint(err) != fmt.Sprint(test.err) {
			t.Errorf("Failed to %v with unexpected error, got:%v expected %v", test.message, err, test.err)
		}
		if !reflect.DeepEqual(got, test.expect) {
			t.Errorf("Failed to %v, got:%v expect:%v", test.message, got, test.expect)
		}
	}
}
