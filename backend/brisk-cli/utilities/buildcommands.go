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

package utilities

import (
	"brisk-supervisor/api"
	"fmt"

	hs "github.com/mitchellh/hashstructure/v2"
)

func HashBuildCommands(buildCommands []*api.Command) (string, error) {
	hash, err := hs.Hash(buildCommands, hs.FormatV2, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(hash), nil

}
