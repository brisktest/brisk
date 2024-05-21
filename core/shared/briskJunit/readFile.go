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

package briskJunit

import (
	"context"
	"sort"

	junit "github.com/joshdk/go-junit"

	. "brisk-supervisor/shared/logger"
)

func ReadFile(ctx context.Context, filename string) ([]junit.Suite, error) {
	suites, err := junit.IngestFile(filename)
	if err != nil {
		Logger(ctx).Debugf("failed to ingest JUnit xml %v", err)
	}
	return suites, err
}

type TestSorter struct {
	FileName string
	Time     int
}

func getFileName(t TestSorter) string {
	return t.FileName
}
func GetTestsByRunTime(ctx context.Context, suites []junit.Suite) []string {

	testRuns := []TestSorter{}
	for _, s := range suites {
		// fmt.Printf("%d : %+v \n", s.Totals.Duration.Milliseconds(), s.Name)
		testRuns = append(testRuns, TestSorter{FileName: s.Name, Time: int(s.Totals.Duration.Milliseconds())})
	}

	sort.Slice(testRuns, func(i, j int) bool {
		return testRuns[i].Time > testRuns[j].Time
	})
	for _, s := range testRuns {
		Logger(ctx).Debugf("%d : %+v \n", s.Time, s.FileName)
	}
	return Map(testRuns, getFileName)
}

func Map(vs []TestSorter, f func(TestSorter) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
