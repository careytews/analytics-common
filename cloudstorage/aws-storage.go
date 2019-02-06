package cloudstorage

import (
	"bytes"
	"errors"
	//"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/trustnetworks/analytics-common/utils"
)

type AWSStorage struct {
	// client                  *http.Client
	bucketName              string
	svc                     *session.Session
	svcoutageretrysleeptime time.Duration // TODO: implement retry around getting object or version info
	bucketRegion            string
}

type AWSGeneration struct {
	Value string
}

func (a *AWSStorage) Init(bucketNameEnvVar string, bucketNameDefault string) {
	a.bucketName = utils.Getenv(bucketNameEnvVar, bucketNameDefault)

	sleepduration, err := strconv.Atoi(utils.Getenv("SVC_OUTAGE_RETRYSLEEPTIME", "10"))
	if err != nil {
		utils.Log("Couldn't get SVC_OUTAGE_RETRYSLEEPTIME: %s,  setting to 10 seconds", err.Error())
		sleepduration = 10
	}

	a.bucketRegion = utils.Getenv("AWS_BUCKET_REGION", "us-west-2")

	// TODO: is this necessary?
	a.svcoutageretrysleeptime = time.Second * time.Duration(int64(sleepduration))
	utils.Log("svcOutageRetrySleepTime set to: %s", a.svcoutageretrysleeptime)

	// We don't need to create a bucket as it is done in provisioning service
	a.createService() // We can do this here as it is not specific for upload/download as it is in GCP.
}

func (a *AWSStorage) createService() {
	// Initialize a session in the region where the bucket is, that the
	// SDK will use to load credentials from the shared credentials file
	// ~/.aws/credentials.
	a.svc = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(a.bucketRegion)},
	))
}

// Multi-part uploads
func (a *AWSStorage) Upload(path string, data []byte) {
	// Upload is AWS-recommended way of storing files (over putObject)
	// Upload function intelligently buffers large files into smaller
	// chunks and sends them in parallel across multiple goroutines.
	// You can use a multipart upload for objects from 5 MB to 5 TB in
	// size (final part may be < 5MB).

	// Setup the S3 Upload Manager.
	// Defaults:
	// - PartSize: 5MB (this is also the minimum size)
	// - Concurrency: 5 (no. of goroutines to spin up in parallel per upload call)
	// - LeavePartsOnError: False (calls AbortMultipartUpload - i.e. removes any
	//	successfully uploaded data)
	// - MaxUploadParts: 10000 (this is also the max)

	uploader := s3manager.NewUploader(a.svc)

	// Upload the file's body to S3 bucket as an object with the key being the
	// same as the filename.
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(a.bucketName),

		// Can also use the `filepath` standard library package to modify the
		// filename as need for an S3 object key. Such as turning absolute path
		// to a relative path.
		Key: aws.String(path),

		// The file to be uploaded. io.ReadSeeker is preferred as the Uploader
		// will be able to optimize memory when uploading large content. io.Reader
		// is supported, but will require buffering of the reader's bytes for
		// each part.
		Body: bytes.NewReader(data),
	})
	if err != nil {
		// Print the error and exit.
		utils.Log("Unable to upload %q to %q, %v", a.bucketName, path, err)
	}
}

func (a *AWSStorage) Download(object string, filepath string, generation CloudGeneration) error {

	gen, ok := generation.(*AWSGeneration)
	if ! ok {
		errStr := "AWSStorage download given none AWS generation"
		utils.Log("ERROR: " + errStr)
		return errors.New(errStr)
	}
	downloader := s3manager.NewDownloader(a.svc)

	file, err := os.Create(filepath)
	if err != nil {
		utils.Log("ERROR: Failed to create file %q, %v", filepath, err)
		return err
	}

	// Download specified version of file
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket:    aws.String(a.bucketName),
			Key:       aws.String(object),
			VersionId: &gen.Value,
		})
	if err != nil {
		utils.Log("ERROR: Couldn't write file: %s", err.Error())
		return err
	}

	return nil
}

func (a *AWSStorage) GetObjectGeneration(object string) CloudGeneration {

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(a.bucketName),
		Prefix: aws.String(object),
	}

	svc := s3.New(a.svc)
	result, err := svc.ListObjectVersions(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				utils.Log("ERROR: Couldn't list object versions: %s", aerr.Code())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			utils.Log("ERROR: Couldn't list object versions: %s", err.Error())
		}
		return nil
	}

	if len(result.Versions) == 0 {
		utils.Log("ERROR: No versions returned for object, maybe the file doesn't exist?")
		return nil
	}

	var version int
	for version = range result.Versions {
		// Find the latest (current) version
		if *result.Versions[version].IsLatest {
			break
		}
	}

	var generation AWSGeneration
	generation.Value = *result.Versions[version].VersionId

	return &generation
}

func (ag *AWSGeneration) Update(value interface{}) error {
	strVal, ok := value.(string)
	if ! ok {
		errStr := "AWSGeneration only accepts string values to update"
		utils.Log("ERROR: " + errStr)
		return errors.New(errStr)
  }
	ag.Value = strVal
  return nil
}

func (ag *AWSGeneration) Equals(rhs CloudGeneration) bool {
	aGen, ok := rhs.(*AWSGeneration)
	// if the other value is not AWS Generation then its not equal
	if ! ok {
		return false
	}
	// compare values
	return aGen.Value == ag.Value
}
