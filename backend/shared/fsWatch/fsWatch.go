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
	. "brisk-supervisor/shared/logger"
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher

var ctx context.Context
var workingDirectory string
var excludedFromWatch []string

// main
func Watch(ctx_in context.Context, root string, controlChan chan string, conf shared.Config) error {
	ctx = ctx_in
	// creates a new file watcher
	workingDirectory = root
	Logger(ctx).Debugf("Working directory is %s", root)
	var err error
	config = conf
	excludedFromWatch = AddFullPathToWatch(ctx, workingDirectory, config.ExcludedFromWatch)

	watcher, err = fsnotify.NewWatcher()
	if err != nil {

		Logger(ctx).Errorf("Error in watcher is %v", err)
		return err
	}
	if watcher == nil {
		Logger(ctx).Fatal("Watcher is nil")
	}
	//defer watcher.Close()

	// starting at the root of the project, walk each file/directory searching for
	// directories
	if err := filepath.WalkDir(root, watchDir); err != nil {
		Logger(ctx).Error("ERROR - %s ", root, err)
		return err
	}
	// for _, v := range excludedFromWatch {
	// 	str := strings.TrimSuffix(v, "/")

	// 	Logger(ctx).Debugf("Removing %v from watch", str)
	// 	watcher.Remove(str)

	// }

	//
	doneWatch := make(chan bool)
	Logger(ctx).Debug("The watchlist is %v", watcher.WatchList())

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				Logger(ctx).Infof("EVENT! %v\n", event)
				if event.String() == "CHMOD" {
					Logger(ctx).Debugf("CHMOD event, ignoring")
				} else {
					if !contains(excludedFromWatch, event.Name) {
						Logger(ctx).Debugf("%v has changed so we need to rerun", event.Name)
						controlChan <- "AutoSave"
					}
				}
				// add to filesChan and then in startSync just sync those files

				// watch for errors
			case err := <-watcher.Errors:
				Logger(ctx).Error("ERROR - ", err)
			}
		}
	}()

	<-doneWatch
	return nil
}
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

var config shared.Config

func AddFullPathToWatch(ctx context.Context, path string, arrayString []string) []string {

	Logger(ctx).Debugf("AddFullPathToWatch called with path %v and arrayString %v", path, arrayString)
	var outArray []string
	for _, v := range arrayString {
		v = strings.TrimSuffix(v, "/")
		outArray = append(outArray, path+"/"+v)
		Logger(ctx).Debugf("Adding %v to excluded watch path", path+"/"+v)
	}
	return outArray
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, entry fs.DirEntry, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	Logger(ctx).Debugf("Looking at %s \n", path)
	//Logger(ctx).Debugf("Looking at %s \n", entry.Name())

	if contains(excludedFromWatch, path) {
		Logger(ctx).Debugf("Excluded from watch %v \n", excludedFromWatch)
		Logger(ctx).Debugf("Not watching %v - excluded in config", path)
		return filepath.SkipDir
	} else {
		Logger(ctx).Debugf("****watching %v - not excluded in config", path)
	}

	if entry.IsDir() && contains(excludedFromWatch, path) {
		Logger(ctx).Debugf("Not watching %v - excluded in config", path)
		return filepath.SkipDir
	}

	if entry.IsDir() {
		//Logger(ctx).Debugf("Watching %s", path)
		if watcher == nil {
			Logger(ctx).Fatal("watcher is nil")
		}

		err := watcher.Add(path)
		Logger(ctx).Debugf("Watched %s", path)

		if err != nil {
			if err.Error() == "bad file descriptor" {
				Logger(ctx).Debugf("Error is '%s' for path '%s'", err, path)
				//return err
			} else {

				Logger(ctx).Debugf("passed bfile Error is '%s' for path '%s'", err, path)
				return err
			}
		}
	}
	return nil

}
