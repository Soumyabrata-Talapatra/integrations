package handler

import (
	"sync"

	"github.com/minio/minio-go/v7"
)

type MinioConfig struct {
	URL        string `mapstructure:"url"`
	AccessKey  string `mapstructure:"accessKey"`
	SecretKey  string `mapstructure:"secretKey"`
	BucketName string `mapstructure:"bucketName"`
}

var (
	once        sync.Once
	MinioClient *minio.Client
	minioConfig MinioConfig
)
