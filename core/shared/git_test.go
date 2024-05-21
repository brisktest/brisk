// Copyright 2024 Brisk, Inc
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

package shared

import (
	"context"
	"testing"
)

// TestHelloEmpty calls greetings.Hello with an empty string,
// checking for an error.
func TestSortStrings(t *testing.T) {

	stringArray := make([]string, 2)
	stringArray[0] = "a"
	stringArray[1] = "z"

	res := SortStrings(stringArray)
	if res[0] != "a" && res[1] != "z" {
		t.Fatalf(`String is not sorted`)
	}
}

func TestHashFileContent(t *testing.T) {

	path := "/tmp/"

	content := HashFileContent(context.Background(), path, "gotest")

	t.Logf(`Hash is %s`, content)

}
