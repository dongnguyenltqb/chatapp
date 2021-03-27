package entity

import (
	"chatapp/infra"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	ImageBucketName = "devteam-s3-test"
)

type Image struct {
	FileType    string `json:"fileType"`
	S3ObjectKey string `json:"s3ObjectKey"`
	FileName    string `json:"fileName"`
}

func (i *Image) GetPreSignedUploadUrl() (string, error) {
	r, _ := infra.GetS3().PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(ImageBucketName),
		Key:         aws.String(i.S3ObjectKey),
		ContentType: aws.String("image/" + i.FileType),
		ACL:         aws.String("public-read"),
	})
	url, err := r.Presign(150 * time.Minute)
	return url, err
}
