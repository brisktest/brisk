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
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CreateTar(dirPath string, tarPath string) error {
	file, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := tar.NewWriter(file)
	defer writer.Close()

	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		header.Name = relPath

		if err := writer.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(writer, file); err != nil {
			return err
		}

		return nil
	})
}

func CreateTarZip(srcPath, destPath string) error {
	// Create the destination file for the zip
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	// Create a zip writer for the destination file
	zipWriter := zip.NewWriter(destFile)
	defer zipWriter.Close()

	// Walk the source directory and add each file to the zip archive
	err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk directory: %v", err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Create a new zip file header for the file
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create file header: %v", err)
		}
		header.Name = path[len(srcPath)+1:]

		// Add the file to the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip header: %v", err)
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("failed to add file to zip archive: %v", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}

	return nil
}
