package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

// GetCephConnection : 获取ceph连接
func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	// 1. 初始化ceph的一些信息

	auth := aws.Auth{
		AccessKey: CephAccessKey,
		SecretKey: CephSecretKey,
	}

	curRegion := aws.Region{
		Name:                 "default",
		EC2Endpoint:          CephGWEndpoint,
		S3Endpoint:           CephGWEndpoint,
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}

	// 2. 创建S3类型的连接
	return s3.New(auth, curRegion)
}

// GetCephBucket : 获取指定的bucket对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}

// PutObject : 上传文件到ceph集群
func PutObject(bucket string, path string, data []byte) error {
	// PublicRead 表示文件为公共读，不需要用户验证（生产环境慎用）
	return GetCephBucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}
