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

package log_service

import (
	"brisk-supervisor/shared/constants"
	"context"
	"fmt"
	"io"
	"time"

	. "brisk-supervisor/shared/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (s *LogService) PrintConfig() {
	fmt.Printf("LogService Config: %+v", s)
}

type LogService struct {
	awsSession   *session.Session
	uploader     *s3manager.Uploader
	bucket       string
	key          string
	awsConfig    *aws.Config
	Location     string
	projectToken string
}

func New(projectToken string, uid string) *LogService {
	// key will identify the specific jobruninfo
	// we can then recombine it with the job run

	ls := LogService{projectToken: projectToken, key: projectToken + "/" + uid}
	ls.Configure()
	return &ls
}

func (s *LogService) Configure() error {

	s.bucket = constants.DEFAULT_LOG_BUCKET
	//load default credentials
	s.awsConfig = &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region:                        aws.String(constants.DEFAULT_REGION),
		LogLevel:                      aws.LogLevel(aws.LogDebug),
	}

	s.awsSession = session.Must(session.NewSession(s.awsConfig))
	s.uploader = s3manager.NewUploader(s.awsSession)

	return nil
}

// pass a reader to the log service and it will upload it to s3

func (s *LogService) UploadContent(ctx context.Context, content io.Reader, logger *BriskLogger) (string, error) {
	logger.Infof("Uploading content to s3 bucket: %s, key: %s", s.bucket, s.key)

	// Upload the content to S3.
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	opts := func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10MB per part
		u.LeavePartsOnError = false   // Don't delete the parts if the upload fails.
		u.Concurrency = 5

	}
	result, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.key),
		Body:        content,
		ContentType: aws.String("text/plain"),
	}, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	logger.Infof("file uploaded to, %s\n", result.Location)
	s.Location = result.Location
	return s.Location, nil
}
