package datasources

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"start/config"
	"start/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

func MinioConnect() {

	var errConnect error
	// Initialize minio client object.
	minioClient, errConnect = minio.New(
		config.GetMinioEndpoint(),
		&minio.Options{
			Creds: credentials.NewStaticV4(
				config.GetMinioAccessKey(),
				config.GetMinioSecretKey(),
				""),
			Secure: config.GetMinioSsl(),
		})
	if errConnect != nil {
		log.Fatalln(errConnect)
	}
	log.Println("minio connected:", config.GetMinioEndpoint())
}

func mackePublicBucket(bucketName string) error {
	// bucketName := config.GetMinioBucket()
	// ImageFolder := "images"

	log.Println("bucketName", bucketName)
	err := minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(context.Background(), bucketName)
		if errBucketExists == nil && exists {
		} else {
			log.Println("minio err mackePublicBucket", err.Error())
			return err
		}

		policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{"Action": ["s3:GetObject"],
			"Effect": "Allow",
			"Principal": {"AWS": ["*"]},
			"Resource": ["arn:aws:s3:::%v/*"]
			,"Sid": "Public"}
			]
		}`
		policy = fmt.Sprintf(policy, bucketName)
		errPolicy := minioClient.SetBucketPolicy(context.Background(), bucketName, policy)
		if errPolicy != nil {
			log.Println("minio err mackePublicBucket policy", errPolicy.Error())
			return errPolicy
		}
	}
	return nil
}

func minioUploadFile(bucketname, path string, file *multipart.File) (string, error) {
	errMake := mackePublicBucket(bucketname)
	if errMake != nil {
		return "", errMake
	}

	var buff bytes.Buffer
	io.Copy(&buff, *file)
	bFile := buff.Bytes()
	lengthFile := buff.Len()
	userMetaData := map[string]string{"x-amz-acl": "public-read"}
	// contentTypeFile := http.DetectContentType(bFile)

	uploadInfo, errUpload := minioClient.PutObject(
		context.Background(),
		bucketname,
		path, bytes.NewReader(bFile),
		int64(lengthFile),
		minio.PutObjectOptions{ContentType: "application/octet-stream", UserMetadata: userMetaData}) // svg contentType : image/svg+xml
	if errUpload != nil {
		log.Println("minio err ERROR_UPLOAD:", errUpload)
		return "", errUpload
	}
	fmt.Println("minio Successfully uploaded bytes: ", utils.StructToJson(uploadInfo))

	return config.GetMinioGetDataHost() + config.GetMinioBucket() + path, nil
}

func UploadAvatarFile(file *multipart.File) (string, error) {
	bucketName := config.GetMinioBucket()
	path := "/avatars/" + utils.GenerateUidTimestamp() + ".png"
	return minioUploadFile(bucketName, path, file)
}

func UploadFile(file *multipart.File) (string, error) {
	bucketName := config.GetMinioBucket()
	path := "/image/" + utils.GenerateUidTimestamp() + ".png"
	return minioUploadFile(bucketName, path, file)
}

func UploadIconFile(iconName, fileType string, file *multipart.File) (string, error) {
	bucketName := config.GetMinioBucket()
	path := "/icon/" + iconName + "-" + utils.GenerateUidTimestamp() + "." + fileType
	return minioUploadFile(bucketName, path, file)
}
