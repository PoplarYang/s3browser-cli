package main

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UserConfig is user basic info
type UserConfig struct {
	EndPoint    string
	Region      string
	BucketName  string
	AccessKeyID string
	SecretKeyID string
	SignedUrlExpiretion time.Duration
}

var userConfig UserConfig
var sess *session.Session

func init() {
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	var err error
	sess, err = session.NewSession(&aws.Config{
		Region:           aws.String(userConfig.Region),
		Endpoint:         aws.String(userConfig.EndPoint),
		Credentials:      credentials.NewStaticCredentials(userConfig.AccessKeyID, userConfig.SecretKeyID, ""),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		log.Fatalf("Authorization failed, %v\n", err)
	}
}

func ListObjects(maxKeys int64, prefix string, startAfter string) (objects []*s3.Object, errors interface{}) {
	// Create S3 service client
	svc := s3.New(sess)
	// Get the list of items
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(userConfig.BucketName),
		MaxKeys:    aws.Int64(maxKeys),
		Prefix:     aws.String(prefix),
		StartAfter: aws.String(startAfter),
	})
	return resp.Contents, err
}

func GetFile(name string) {
	// Create S3 service client
	svc := s3.New(sess)

	object, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(userConfig.BucketName),
		Key:    aws.String(name),
	})

	if err != nil {
		log.Fatalf("Get file failed, %v\n", err)
	}

	if strings.Contains(name, "/") {
		names := strings.Split(name, "/")
		name = names[len(names)-1]
	}

	newFile, err := os.Create(name)
	numBytesWritten, err := io.Copy(newFile, object.Body)
	defer object.Body.Close()
	defer newFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Downloaded %d byte file.\n", numBytesWritten)
}

// refer https://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3manager/#NewDownloader
func DownloadObject(filePath string) {
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	fileName := filePath
	// process if path has "/"
	if strings.Contains(filePath, "/") {
		names := strings.Split(filePath, "/")
		fileName = names[len(names) - 1]
	}

	// Create a file to write the S3 Object contents to.
	fd, err := os.Create(fileName)
	defer fd.Close()

	if err != nil {
		log.Fatalf("Fail to create file %q, %v", filePath, err)
		os.Exit(1)
	}

	// Write the contents of S3 Object to the file
	n, err := downloader.Download(fd, &s3.GetObjectInput{
		Bucket: aws.String(userConfig.BucketName),
		Key:    aws.String(filePath),
	})

	if err != nil {
		log.Fatalf("Fail to download file, %v", err)
	}
	log.Printf("File downloaded, %d bytes\n", n)
}

func StatFile(name string) *s3.HeadObjectOutput {
	// Create S3 service client
	svc := s3.New(sess)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(userConfig.BucketName),
		Key:    aws.String(name),
	}

	result, err := svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Fatalf(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Fatalf(err.Error())
		}
		return nil
	}
	return result
}

func SignS3Url(name string) {
	// Create S3 service client
	svc := s3.New(sess)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(userConfig.BucketName),
		Key:    aws.String(name),
	})
	url, err := req.Presign(userConfig.SignedUrlExpiretion)
	if err != nil {
		log.Fatalf("Get signed url failed, %v", err)
	}

	log.Printf("Signed url is: %v\n", url)
}

func SetObjectPublicRead(name string) {
	// Create S3 service client
	svc := s3.New(sess)

	objectAclInput := &s3.PutObjectAclInput{
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
		Bucket: aws.String(userConfig.BucketName),
		Key:    aws.String(name),
	}

	_, err := svc.PutObjectAcl(objectAclInput)
	if err != nil {
		log.Fatalf("Set object ACL failed, %v", err)
	}
}

