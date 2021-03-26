package infra

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var onceInitS3 sync.Once
var s3Svc *s3.S3

func GetS3() *s3.S3 {
	onceInitS3.Do(func() {
		sess, err := session.NewSession(&aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        aws.String("ap-southeast-1")},
		)
		if err != nil {
			panic(err)
		}
		// Create S3 service client
		s3Svc = s3.New(sess)

	})
	return s3Svc
}
