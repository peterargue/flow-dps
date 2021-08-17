// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package bucket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/optakt/flow-dps/models/dps"
	"github.com/rs/zerolog"
)

// FIXME: Switch to GCP bucket.

type Downloader struct {
	logger zerolog.Logger

	// bucket is the name of the S3 bucket to use.
	bucket string

	// service is used to regularly poll the AWS API to check for new items in the bucket.
	service *s3.S3

	// downloader is used to download bucket items into the filesystem.
	downloader *s3manager.Downloader

	// buffer stores the bytes of the current segment until it is read, at which point it gets reset.
	buffer *aws.WriteAtBuffer

	// segment holds a copy of the bytes from buffer, but gets dynamically truncated as it is read.
	segment []byte

	// the index of the next segment to retrieve.
	index int

	wg      sync.WaitGroup
	stopCh  chan struct{}
	readCh  chan struct{}
	writeCh chan struct{}
}

// NewDownloader creates a new AWS S3 session for downloading and returns a Downloader instance that uses it.
func NewDownloader(logger zerolog.Logger, region, bucket string) (*Downloader, error) {
	cfg := &aws.Config{
		Region: aws.String(region),
	}

	// Create the S3 session used to list bucket items and download them.
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not initialize AWS session: %w", err)
	}

	d := Downloader{
		logger:     logger,
		bucket:     bucket,
		service:    s3.New(sess),
		downloader: s3manager.NewDownloader(sess),
		buffer:     &aws.WriteAtBuffer{},
		stopCh:     make(chan struct{}),
		readCh:     make(chan struct{}),
		writeCh:    make(chan struct{}),
	}

	return &d, nil
}

// Next reads and returns the contents of the next available segment.
func (d *Downloader) Next() ([]byte, error) {
	select {
	case <-d.writeCh:
		// Proceed with reading the file.
	case <-time.After(5 * time.Second):
		// Time out, since no file is available at the moment.
		return nil, dps.ErrTimeout
	}

	// Copy the contents of the current buffer, reset it and let the reading
	// goroutine know that it can write over it again.
	data := d.buffer.Bytes()
	d.buffer = &aws.WriteAtBuffer{}
	d.readCh <- struct{}{}

	return data, nil
}

// Run is designed to run in the background and download the next segment whenever one has been fully read.
func (d *Downloader) Run() {
	d.wg.Add(1)
	for {
		// Check if we should stop.
		select {
		case <-d.stopCh:
			return
		default:
		}

		// List bucket items.
		req := &s3.ListObjectsV2Input{Bucket: aws.String(d.bucket)}
		resp, err := d.service.ListObjectsV2(req)
		if err != nil {
			d.logger.Error().Err(err).Msg("could not list bucket items")
			continue
		}

		// Segments are formatted like: 00000000, 00000001, etc.
		wantKey := fmt.Sprintf("%08d", d.index)

		// Look at each bucket item to check whether or not it has already been downloaded. If not, download it.
		var totalDownloadedBytes int64
		for _, item := range resp.Contents {
			// Skip any item that does not match the index we are looking for.
			if *item.Key != wantKey {
				continue
			}

			// Download the segment into the buffer.
			req := &s3.GetObjectInput{
				Bucket: aws.String(d.bucket),
				Key:    aws.String(*item.Key),
			}
			n, err := d.downloader.Download(d.buffer, req)
			if err != nil {
				d.logger.Error().Err(err).Msg("could not download bucket item")
				continue
			}

			totalDownloadedBytes += n
		}

		if totalDownloadedBytes == 0 {
			d.logger.Debug().
				Str("bucket", d.bucket).
				Str("segment_key", wantKey).
				Int64("download_bytes", totalDownloadedBytes).
				Msg("waiting for segment to be available")

			// FIXME: Seems necessary to avoid spamming the S3 API.
			//		  Set to a low value for now for testing.
			// 		  How often should we poll realistically?
			// 		  Should this be configurable?
			time.Sleep(1 * time.Second)
		} else {
			d.logger.Debug().
				Str("bucket", d.bucket).
				Str("segment_key", wantKey).
				Int64("download_bytes", totalDownloadedBytes).
				Msg("successfully downloaded segment from bucket")

			// Notify that the buffer is ready to be read.
			d.writeCh <- struct{}{}

			// Wait until buffer has been read and reset by the feeder before
			// downloading the next one.
			<-d.readCh
		}
	}
}

// Stop makes the Run and Read methods stop gracefully.
func (d *Downloader) Stop(ctx context.Context) error {
	close(d.stopCh)
	close(d.readCh)
	d.wg.Wait()
	return nil
}
