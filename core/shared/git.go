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
	api "brisk-supervisor/api"
	. "brisk-supervisor/shared/logger"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	git "github.com/go-git/go-git/v5"
	"golang.org/x/mod/sumdb/dirhash"
)

// take a list of files and hash the contents

func SortStrings(strings []string) []string {
	sort.Strings(strings)
	// fmt.Println(strings)

	return strings
}

// check that the files exist and are readable
func CheckFilesExist(ctx context.Context, workingDirectory string, files []string) error {
	Logger(ctx).Debugf("Checking if files exist %s with workingDirectory %v", files, workingDirectory)
	for _, file := range files {
		if _, err := os.Stat(workingDirectory + file); os.IsNotExist(err) {
			Logger(ctx).Debug("Error in file check")
			Logger(ctx).Debug("Error in file check with workingDirectory " + workingDirectory)
			Logger(ctx).Debug("Before Error")

			Logger(ctx).Errorf("File %s does not exist - %v", workingDirectory+file, err)
			Logger(ctx).Error("after error")
			Logger(ctx).Debug("After Error")
			return fmt.Errorf("file does not exist :  %s", file)
		}
	}
	return nil
}

func HashFiles(ctx context.Context, hashFilePath string, workingDirectory string, files []string) (string, error) {
	Logger(ctx).Debugf("Hashing files %s with workingDirectory %v and hasFilePath %v", files, workingDirectory, hashFilePath)
	fileErr := CheckFilesExist(ctx, workingDirectory, files)
	if fileErr != nil {

		Logger(ctx).Errorf("Error checking files : %v", fileErr)
		return "", fileErr
	}

	var content string

	for _, file := range SortStrings(files) {
		Logger(ctx).Debugf("Hashing file %s", file)
		c := HashFileContent(ctx, workingDirectory, file)
		Logger(ctx).Debugf("content of file is %s", c)

		content += c
	}
	// fmt.Println(content)

	output := fmt.Sprintf("%x", HashString(ctx, content))

	writeErr := WriteHash(ctx, []byte(output), hashFilePath)
	if writeErr != nil {
		Logger(ctx).Errorf("Error writing hash to file : %v", writeErr)
		return "", writeErr
	}

	return output, nil

}

func HashString(ctx context.Context, s string) string {
	sha256 := sha256.New()
	_, err := sha256.Write([]byte(s))
	if err != nil {

		SafeExit(err)

	}
	return string(sha256.Sum(nil))
}

func HashFileContent(ctx context.Context, workingDirectory string, f string) string {

	file, err := os.Open(workingDirectory + f)
	if err != nil {
		SafeExit(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		SafeExit(err)
	}
	if fileInfo.IsDir() {
		Logger(ctx).Debugf("File %s is a directory", workingDirectory+f)
		hash, err := dirhash.HashDir(workingDirectory+f, f, dirhash.DefaultHash)

		if err != nil {
			SafeExit(err)
		}
		Logger(ctx).Debugf("Hash of directory %s is %s", workingDirectory+f, hash)
		return hash
	}

	defer file.Close()

	buf := make([]byte, 30*1024)
	sha256 := sha256.New()
	for {
		n, err := file.Read(buf)
		if n > 0 {
			_, err := sha256.Write(buf[:n])
			if err != nil {
				SafeExit(err)
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			Logger(ctx).Debugf("Read %d bytes: %v", n, err)
			break
		}
	}

	sum := sha256.Sum(nil)
	return string(sum)
}

func EnsureFileExists(ctx context.Context, file string) error {
	if len(file) == 0 {

		return errors.New("no file provided ")
	}
	var err error
	if _, err = os.Stat(file); os.IsNotExist(err) {
		_, err = os.Create(file)
		if err != nil {
			SafeExit(err)
		}
	}
	return err
}

func WriteHash(ctx context.Context, hash []byte, file string) error {

	//we are deliberatley truncating the file here to ensure that we are overwriting the file
	if len(file) == 0 {
		return errors.New("no file path provided ")
	}

	f, fileErr := os.Create(file)
	if fileErr != nil {
		Logger(ctx).Debug(fileErr)
		return fileErr
	}

	_, err := f.Write([]byte(hash))
	if err != nil {
		SafeExit(err)
	}
	Logger(ctx).Debugf("wrote hash to file %s", f.Name())
	return nil
}

func ReadRebuildHash(ctx context.Context, hashFilePath string) (string, error) {
	return ReadFile(ctx, hashFilePath)

}

func RebuildRequired(ctx context.Context, hashFilePath string, workingDirectory string, files []string) (bool, error) {
	Logger(ctx).Debugf("Checking if rebuild is required for files %s with workingDirectory %v and hasFilePath %v", files, workingDirectory, hashFilePath)
	storedHash, err := ReadRebuildHash(ctx, hashFilePath)
	if err != nil {
		Logger(ctx).Errorf("Error reading hash file %s - %v", hashFilePath, err)
		return false, err
	}

	newHash, hashErr := HashFiles(ctx, hashFilePath, workingDirectory, files)

	if hashErr != nil {
		Logger(ctx).Errorf("getting a hash error %v", hashErr)
		return false, hashErr
	}

	if len(storedHash) == 0 {
		Logger(ctx).Debugf("No hash stored, no rebuild required")
		return false, nil
	}

	Logger(ctx).Debugf("Stored hash ==  %s new hash == %s  ", storedHash, newHash)

	return storedHash != newHash, nil

}

func ReadFile(ctx context.Context, filePath string) (string, error) {
	if len(filePath) == 0 {
		return "", errors.New("no file path provided ")
	}
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		Logger(ctx).Debugf("Error opening file %s : - %v", filePath, err)

		return "", err
	}
	defer f.Close()
	var content string
	buf := make([]byte, 30*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			content += string(buf[:n])
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			Logger(ctx).Debugf("Read %d bytes: %v", n, err)
			break
		}
	}
	Logger(ctx).Debugf("Read file %s : %s", filePath, content)
	return content, nil
}

func CheckIfGitRepo(ctx context.Context, dirPath string) (bool, error) {
	if len(dirPath) == 0 {
		return false, errors.New("no directory path provided ")
	}
	//check if  path is a git repo
	_, err := os.Stat(dirPath)
	if err != nil {
		Logger(ctx).Debug(err)
		return false, err
	}

	_, err = os.Stat(dirPath + "/.git")
	if err != nil {
		Logger(ctx).Debug(err)
		return false, err
	}

	return true, nil

}

// get git commit and branch name
func GetGitCommitAndBranch(ctx context.Context, dirPath string) (string, string, error) {
	if len(dirPath) == 0 {
		return "", "", errors.New("no directory path provided ")
	}
	//check if  path is a git repo
	isGitRepo, err := CheckIfGitRepo(ctx, dirPath)
	if err != nil {
		Logger(ctx).Debug(err)
		return "", "", err
	}

	if !isGitRepo {
		return "", "", errors.New("not a git repo")
	}

	//get git commit
	commit, message, err := GetGitCommit(ctx, dirPath)
	if err != nil {
		Logger(ctx).Debug(err)
		return "", "", err
	}

	// //get git branch
	// branch, err := GetGitBranch(ctx, dirPath)
	// if err != nil {
	// 	Logger(ctx).Debug(err)
	// 	return "", "", err
	// }

	return commit, message, nil
}

func GetRemote(ctx context.Context, repo *git.Repository) (*git.Remote, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		Logger(ctx).Debug(err)
		return nil, err
	}

	if len(remotes) == 0 {
		return nil, errors.New("no remotes found")
	}

	return remotes[0], nil

}

func GetRemoteForPath(ctx context.Context, path string) (string, error) {

	repo, err := git.PlainOpen(path)
	if err != nil {
		Logger(ctx).Debug(err)
		return "", err
	}

	remote, err := GetRemote(ctx, repo)
	if err != nil {
		Logger(ctx).Debug(err)
		return remote.String(), err
	}

	Logger(ctx).Debugf("remote url %s", remote.Config().URLs[0])
	return remote.Config().URLs[0], nil
}

// func GetRepoName()

//	message RepoInfo {
//		string CommitHash = 1;
//		string Repo = 2;
//		string Branch = 3;
//		string Tag = 4;
//		string CommitMessage = 5;
//		string CommitAuthor = 6;
//		string CommitAuthorEmail = 7;
//		bool IsGitRepo = 8;
//	  }
func GetGitInfo(ctx context.Context, path string) (*api.RepoInfo, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		Logger(ctx).Debug(err)
		return &api.RepoInfo{IsGitRepo: false}, err
	}

	head, err := repo.Head()
	if err != nil {
		Logger(ctx).Debug(err)
		return &api.RepoInfo{IsGitRepo: false}, err
	}

	commit, err := repo.CommitObject(head.Hash())

	return &api.RepoInfo{Branch: string(head.Name()), CommitHash: commit.Hash.String(), CommitMessage: commit.Message, CommitAuthor: commit.Author.Name, CommitAuthorEmail: commit.Author.Email, IsGitRepo: true}, nil
}

// get git commit
func GetGitCommit(ctx context.Context, dirPath string) (string, string, error) {

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		Logger(ctx).Debug(err)
		return "", "", err
	}

	head, err := repo.Head()
	if err != nil {
		Logger(ctx).Debug(err)
		return "", "", err
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		Logger(ctx).Debug(err)
		return "", "", err
	}

	return commit.Hash.String(), commit.Message, nil

}
