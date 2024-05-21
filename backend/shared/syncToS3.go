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
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func SyncToS3(bucket string, key string, accessKey string, secretKey string, tarPath string) error {
	// Parse the S3 URL

	// Get the content type of the tar file
	contentType := "application/x-tar"

	signedURL := generateS3PUTURL("PUT", contentType, key, 15*time.Minute, bucket, accessKey, secretKey)
	// Read the tar file into a buffer
	tarData, err := ioutil.ReadFile(tarPath)
	if err != nil {
		return err
	}
	// Create an HTTP request to upload the tar file to S3
	req, err := http.NewRequest("PUT", signedURL, bytes.NewReader(tarData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	currentTime := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", currentTime)

	// Make the HTTP request to upload the tar file to S3
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to sync to S3, status code: %d", resp.StatusCode)
	}

	return nil
}

func parseS3URL(s3URL string) (bucket, key string, err error) {
	// Parse the S3 URL
	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		return "", "", err
	}

	// Get the bucket and key from the URL path
	parts := strings.SplitN(strings.Trim(parsedURL.Path, "/"), "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid S3 URL: %s", s3URL)
	}
	return parts[0], parts[1], nil
}

func generateS3PUTURL(httpMethod, contentType, objectKey string, expireAt time.Duration, bucketName, accessKey, secretKey string) string {
	// Create an AWS session with the provided credentials
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))

	// Create an S3 client
	svc := s3.New(sess)

	// Create a signed URL for the S3 object
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	})
	urlStr, err := req.Presign(expireAt)
	if err != nil {
		fmt.Println("Failed to sign request", err)
	}

	// Return the signed URL as a string
	return urlStr
}

func generateSignedGetURL(httpMethod, contentType, objectKey string, expireAt time.Duration, bucketName, accessKey, secretKey string) string {
	// Create an AWS session with the provided credentials
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))

	// Create an S3 client
	svc := s3.New(sess)

	// Create a signed URL for the S3 object
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	urlStr, err := req.Presign(time.Duration(expireAt))
	if err != nil {
		fmt.Println("Failed to sign request", err)
	}

	// Set the content type for the signed URL
	u, _ := url.Parse(urlStr)
	q := u.Query()
	q.Set("content-type", contentType)
	u.RawQuery = q.Encode()

	// Return the signed URL as a string
	return u.String()
}

func UntarFromS3(s3URL, accessKey, secretKey, destPath string) error {
	// Parse the S3 URL to get the bucket name and object key
	bucket, key, err := parseS3URL(s3URL)
	if err != nil {
		return err
	}

	// Generate a signed URL for downloading the tar file from S3
	expirationTime := 15 * time.Minute
	contentType := "application/x-tar"
	signedURL := generateSignedGetURL("GET", contentType, key, expirationTime, bucket, accessKey, secretKey)

	// Download the tar file from S3 using the signed URL
	resp, err := http.Get(signedURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download tar file, status code: %d", resp.StatusCode)
	}

	// Read the tar file from the HTTP response body into a buffer
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return err
	}

	// Extract the tar file to the destination directory
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Get the destination path for the file
		filePath := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		}
	}

	return nil
}
