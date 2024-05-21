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

package fsWatch

import (
	"brisk-supervisor/shared"
	"context"
	"fmt"
	"os"
	"testing"
)

// makes a directory
func MakeDirectory(path string) error {
	err := os.Mkdir(path, 0755)
	return err

}
func setup() {
	// creates a directory
	p := "/tmp/gotest"
	err := MakeDirectory(p)
	if err != nil {
		fmt.Printf(`Error in creating directory %v`, err)
	}

	err = MakeDirectory(p + "/.git")
	if err != nil {
		fmt.Printf(`Error in creating directory %v`, err)
	}

	err = MakeDirectory(p + "/.git/objects")
	if err != nil {
		fmt.Printf(`Error in creating directory %v`, err)
	}
	_, ferr := os.Create(p + "/.git/objects/1")
	if ferr != nil {
		fmt.Printf(`Error in creating file %v`, err)
	}
}

func shutdown() {
	cleanup("/tmp/gotest")
}

func TestMain(m *testing.M) {
	setup()
	defer shutdown()
	code := m.Run()

	os.Exit(code)
}

// tests the watch function
func TestWatch(t *testing.T) {

	controlChan := make(chan string)

	go func() {
		err := Watch(context.Background(), "/tmp/gotest", controlChan, shared.Config{})
		if err != nil {
			t.Fatalf(`Error in creating directory %v`, err)
		}
	}()
	go func() {
		for {
			out := <-controlChan
			t.Logf(`Control channel output is %v`, out)
		}
	}()

}

func cleanup(path string) {
	// removes the directory
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf(`Error in removing directory %v`, err)

	}
}
