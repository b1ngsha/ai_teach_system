package services

import (
	"ai_teach_system/config"
	"fmt"
	"mime/multipart"
	"path"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSSServiceInterface interface {
	UploadAvatar(file *multipart.FileHeader) (string, error)
}

type OSSService struct {
	client *oss.Client
	bucket *oss.Bucket
}

func NewOSSService() (*OSSService, error) {
	client, err := oss.New(
		config.OSS.Endpoint,
		config.OSS.AccessKeyID,
		config.OSS.AccessKeySecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %v", err)
	}

	bucket, err := client.Bucket(config.OSS.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %v", err)
	}

	return &OSSService{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *OSSService) UploadAvatar(file *multipart.FileHeader) (string, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 生成OSS对象名称
	ext := path.Ext(file.Filename)
	objectKey := fmt.Sprintf("avatars_%d%s", time.Now().Unix(), ext)

	// 上传到OSS
	err = s.bucket.PutObject(objectKey, src)
	if err != nil {
		return "", err
	}

	// 返回可访问的URL
	return fmt.Sprintf("https://%s.%s/%s", config.OSS.BucketName, config.OSS.Endpoint, objectKey), nil
}
