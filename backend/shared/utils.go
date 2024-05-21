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
	pb "brisk-supervisor/brisk-supervisor"
	"crypto/rand"
	"fmt"
)

func PrintWorkerDetails(output *pb.Output) string {
	if output.Worker != nil {
		// return fmt.Sprintf("%v %v", output.Worker.Number, output.Worker.Uid)
		return fmt.Sprintf(output.Worker.Uid)
	}
	return ""
}

func Get32ByteKey() (string, error) {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return string(key), nil
}
