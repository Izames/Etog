package s3

import (
	"Etog/internal/config"
	"strings"

	"mime/multipart"
	neturl "net/url" // алиас, т.к. переменная тоже называется url
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

type S3 struct {
	session *session.Session
	config  *config.S3
}

func NewS3(conf *config.S3) (*S3, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(conf.Region),
		Endpoint:    aws.String(conf.Endpoint),
		Credentials: credentials.NewStaticCredentials(conf.Access, conf.Secret, ""),
	}))
	return &S3{
		session: sess,
		config:  conf,
	}, nil
}

func (s *S3) Upload(fileHeader *multipart.FileHeader, pathS3 string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}

	defer file.Close()

	ext := path.Ext(fileHeader.Filename)
	name := fileHeader.Filename[:len(fileHeader.Filename)-len(ext)]
	safeFilename := name + "_" + uuid.NewString() + ext

	uploader := s3manager.NewUploader(s.session)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(pathS3 + safeFilename),
		Body:   file,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", err
	}
	return result.Location, nil
}

func (s *S3) Delete(url string) error {
	svc := s3.New(s.session)

	parsed, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	key := strings.TrimPrefix(parsed.Path, "/"+s.config.Bucket+"/")

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	return err
}
