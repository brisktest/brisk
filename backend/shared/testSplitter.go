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
	"brisk-supervisor/api"
	pb "brisk-supervisor/brisk-supervisor"
	junit "brisk-supervisor/shared/briskJunit"
	. "brisk-supervisor/shared/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	. "brisk-supervisor/shared/context"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func lsFile(filepath string) {
	ctx := context.TODO()
	Logger(ctx).Debugf("in the lsFile for %s", filepath)
	cmd := exec.Command("bash", "-i", "-c", fmt.Sprintf("ls -a %s", filepath))

	out, err := cmd.Output()
	if err != nil {
		Logger(ctx).Debug(err)
	}

	Logger(ctx).Debug(string(out))
}

func getFileSize(filepath string) (int64, error) {
	ctx := context.TODO()

	//Logger(ctx).Debug("gunna do the remote_dir")
	//lsFile("/tmp/remote_dir")
	//Logger(ctx).Debug("Now doing the filepath")
	//lsFile(filepath)

	fi, err := os.Stat(filepath)
	if err != nil {
		Logger(ctx).Debug(err)
		Logger(ctx).Errorf("Could not obtain stat for file %s", filepath)
		return 0, err
	}

	return fi.Size(), nil
}

func GetAllFiles(ctx context.Context, outputStream chan pb.Output, workDir string) ([]string, error) {
	Logger(ctx).Debug("Getting files with jest")
	listTestCmd := "yarn test --listTests --json"

	cmd := exec.CommandContext(ctx, "bash", "-i", "-c", "-l", listTestCmd)

	env := os.Environ()
	cmd.Env = env
	cmd.Dir = workDir
	Logger(ctx).Debug("About to run Jest listTests")
	out, err := cmd.CombinedOutput()
	Logger(ctx).Debug(string(out))
	outputStream <- pb.Output{Response: string(out)}

	if err != nil {
		outputStream <- pb.Output{Response: cmd.String()}
		Logger(ctx).Errorf("Error running command when getting files %s", cmd.String())
		return nil, err
	}

	split := strings.Split(string(out), "\n")
	fileList := split[len(split)-3]
	Logger(ctx).Debugf("The fileList is %s", fileList)

	dataJson := fileList
	var arr []string
	jsonErr := json.Unmarshal([]byte(dataJson), &arr)
	if jsonErr != nil {
		Logger(ctx).Error(jsonErr)
		return nil, jsonErr
	}
	Logger(ctx).Debugf("Unmarshaled: %v", arr)

	files := arr
	Logger(ctx).Debugf("All files are %v", files)
	Logger(ctx).Debugf("we have %v files ", len(files))

	return files, nil
}

func NewRunACommand(ctx context.Context, command api.Command) (output pb.Output, err error) {
	Logger(ctx).Debugf("NewRunACommand: %v", command)
	cmd := exec.CommandContext(ctx, "bash", "-i", "-c", "-l", command.Commandline+" "+strings.Join(command.Args, " "))

	env := os.Environ()
	cmd.Env = env
	cmd.Dir = command.WorkDirectory
	Logger(ctx).Debug(cmd.String())
	Logger(ctx).Debugf("The work directory is %v", cmd.Dir)
	Logger(ctx).Debug("about to run the command")
	out, err := cmd.Output()
	Logger(ctx).Debug("ran the command")
	Logger(ctx).Debugf("the error is %v", err)

	Logger(ctx).Debug(string(GetLastNBytes(1000, out)))

	Logger(ctx).Debug("^^^^^^^^^^^^^^^^^^^^^")
	if err != nil {

		Logger(ctx).Debug(err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			Logger(ctx).Debug(string(exiterr.Stderr))
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				Logger(ctx).Debugf("Exit Status: %d", status.ExitStatus())

				Logger(ctx).Debug(err.Error())
				output = pb.Output{Stdout: string(out), Stderr: string(exiterr.Stderr), Exitcode: int32(status.ExitStatus()), BriskError: &pb.BriskError{Fatal: true, AdditionalMessage: "Failed running one off command"}, Created: timestamppb.Now()}
				return output, err
			}
		}

		return output, err
	}
	output = pb.Output{Stdout: string(out)}
	return output, err
}

func RunACommand(ctx context.Context, command string, workDir string) error {
	cmd := exec.CommandContext(ctx, "bash", "-i", "-c", "-l", command)
	responseStream, ctxErr := ResponseStream(ctx)
	if ctxErr != nil {
		Logger(ctx).Debug("Error getting the context")
		Logger(ctx).Debug(ctxErr)
		return ctxErr
	}

	env := os.Environ()
	cmd.Env = env
	cmd.Dir = workDir
	Logger(ctx).Debug(cmd.String())
	Logger(ctx).Debug("about to run the command")
	out, err := cmd.CombinedOutput()

	if err != nil {
		*responseStream <- pb.Output{Response: string(out), Stderr: string(out), Created: timestamppb.Now()}
		*responseStream <- pb.Output{Response: "Error - quitting run", Created: timestamppb.Now()}
		return err
	}
	*responseStream <- pb.Output{Stdout: string(out), Created: timestamppb.Now()}

	return err

}

// FileSize describes a test file and it's size
type FileSize struct {
	size       int64
	path       string
	RemotePath string
}

func getFileSizes(ctx context.Context, workDir string, pattern string, remoteDirectory string, outputSteam chan pb.Output) ([]FileSize, error) {
	files, err := GetAllFiles(ctx, outputSteam, workDir)
	if err != nil {
		return nil, err
	}
	var data = make([]FileSize, len(files))
	for i, file := range files {
		var lastSize int64
		if i == 0 {
			lastSize = 0
		} else {
			lastSize = data[i-1].size
		}
		lsFile("/tmp/remote_dir")
		Logger(ctx).Debug("now getFileSize")
		getFileSize(file)
		Logger(ctx).Debug("after first getFileSize")
		fileSize, err := getFileSize(file)
		if err != nil {
			Logger(ctx).Errorf("Error getting file size %s", file)
			return nil, err
		}
		fileSize += lastSize
		data[i] = FileSize{size: fileSize, path: file, RemotePath: strings.Replace(file, workDir, "", 1)}
	}

	return data, nil
}

func GetFilesByJunitJest(ctx context.Context, filename string) ([]string, error) {
	suites, err := junit.ReadFile(ctx, filename)
	results := junit.GetTestsByRunTime(ctx, suites)
	return results, err
}

// SplitFilesByJest divides the files into n buckets based on the time they take (that jest calculates)
func SplitFilesByJest(ctx context.Context, workDir string, num_buckets int, outputStream chan *pb.Output, files []string) ([][]string, int, error) {
	Logger(ctx).Debugf("In SplitFilesByJest+")

	result := make([][]string, num_buckets)
	for i := 0; i < num_buckets; i++ {
		for j := i; j < len(files); j = j + num_buckets {
			localPath := files[j]
			remotePath := strings.Replace(localPath, workDir, "", 1)
			result[i] = append(result[i], string(remotePath))
		}
	}
	return result, len(files), nil
}

// SplitFilesBySize take a pattern and returns an array with the worker files for each worker
func SplitFilesBySize(ctx context.Context, workDir string, pattern string, numWorkers int, remoteDirectory string, outputStream chan pb.Output) ([][]FileSize, int, error) {
	cumulativeSizes, err := getFileSizes(ctx, workDir, pattern, remoteDirectory, outputStream)
	if err != nil {
		return nil, 0, err
	}
	filesForWorkers := make([][]FileSize, len(cumulativeSizes))

	Logger(ctx).Debug(cumulativeSizes)
	if len(cumulativeSizes) == 0 {
		Logger(ctx).Error("No files found")
		return nil, 0, errors.New("no files found")
	}
	maxSize := cumulativeSizes[len(cumulativeSizes)-1].size
	eachSize := maxSize / int64(numWorkers)

	prevIndex := 0
	for n := 0; n < numWorkers; n++ {
		lastIndex := prevIndex
		mySize := (int64(n) + 1) * eachSize
		for ; cumulativeSizes[lastIndex].size < mySize; lastIndex++ {
		}
		filesForWorkers[n] = cumulativeSizes[prevIndex:lastIndex]
		prevIndex = lastIndex

	}
	return filesForWorkers, len(cumulativeSizes), nil

}

type PreTestInfo struct {
	Filenames      []string
	TotalTestCount int
	TotalSkipCount int
}
type JestFiles struct {
	Files []string
}

func SplitForJest(ctx context.Context, jsonString string) (PreTestInfo, error) {
	Logger(ctx).Debug("splitForJest++")
	Logger(ctx).Debugf("the json string is %v", jsonString)
	var result JestFiles

	err := json.Unmarshal([]byte(jsonString), &result.Files)
	if err != nil {
		Logger(ctx).Errorf("Error parsing json %v : the string is %v", err, jsonString)

		return PreTestInfo{}, fmt.Errorf("error unmarshaling string '%v', error: %v", jsonString, err)
	}
	var fileNames []string
	Logger(ctx).Debugf("the result.Examples length is %v", len(result.Files))
	for i := 0; i < len(result.Files); i++ {
		fileNames = append(fileNames, result.Files[i])

	}

	Logger(ctx).Debug("splitForJest--")
	return PreTestInfo{Filenames: RemoveDuplicatesFromSlice(ctx, fileNames), TotalTestCount: len(result.Files), TotalSkipCount: 0}, nil
}

func SplitForRSpec(ctx context.Context, jsonString string) (PreTestInfo, error) {
	Logger(ctx).Debug("splitForRSpec++")
	Logger(ctx).Debugf("the json string is %v", jsonString)
	var result RspecFiles

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		Logger(ctx).Debugf("Error parsing json %v", err)
		return PreTestInfo{}, err
	}
	var fileNames []string
	Logger(ctx).Debugf("the result.Examples length is %v", len(result.Examples))
	for i := 0; i < len(result.Examples); i++ {
		fileNames = append(fileNames, result.Examples[i].FilePath)

	}
	Logger(ctx).Info("splitForRSpec--")
	Logger(ctx).Debug("splitForRSpec--")
	return PreTestInfo{Filenames: RemoveDuplicatesFromSlice(ctx, fileNames), TotalTestCount: result.Summary.ExampleCount, TotalSkipCount: result.Summary.PendingCount}, nil
}

type RspecFiles struct {
	Version  string `json:"version"`
	Seed     int    `json:"seed"`
	Examples []struct {
		ID              string      `json:"id"`
		Description     string      `json:"description"`
		FullDescription string      `json:"full_description"`
		Status          string      `json:"status"`
		FilePath        string      `json:"file_path"`
		LineNumber      int         `json:"line_number"`
		RunTime         float64     `json:"run_time"`
		PendingMessage  interface{} `json:"pending_message"`
	} `json:"examples"`
	Summary struct {
		Duration                     float64 `json:"duration"`
		ExampleCount                 int     `json:"example_count"`
		FailureCount                 int     `json:"failure_count"`
		PendingCount                 int     `json:"pending_count"`
		ErrorsOutsideOfExamplesCount int     `json:"errors_outside_of_examples_count"`
	} `json:"summary"`
	SummaryLine string `json:"summary_line"`
}

func RemoveDuplicatesFromSlice(ctx context.Context, s []string) []string {
	Logger(ctx).Debug("RemoveDuplicatesFromSlice++")
	m := make(map[string]bool)
	for _, item := range s {
		m[item] = true
	}

	var result []string
	for item := range m {
		result = append(result, item)
	}
	Logger(ctx).Debug("RemoveDuplicatesFromSlice--")
	return result
}
