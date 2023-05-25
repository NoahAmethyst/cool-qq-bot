package file_util

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tencentyun/cos-go-sdk-v5"

	"net/http"
	"net/url"
	"os"
)

const (
	bucketUrl     = "https://bot-1317156498.cos.ap-nanjing.myqcloud.com"
	getServiceUrl = "https://cos.COS_REGION.myqcloud.com"
)

var cli *cos.Client

func init() {
	// 存储桶名称，由 bucketname-appid 组成，appid 必须填入，可以在 COS 控制台查看存储桶名称。 https://console.cloud.tencent.com/cos5/bucket
	// 替换为用户的 region，存储桶 region 可以在 COS 控制台“存储桶概览”查看 https://console.cloud.tencent.com/ ，关于地域的详情见 https://cloud.tencent.com/document/product/436/6224 。
	u, _ := url.Parse(bucketUrl)
	// 用于 Get Service 查询，默认全地域 service.cos.myqcloud.com
	su, _ := url.Parse(getServiceUrl)
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	cli = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			// 通过环境变量获取密钥
			// 环境变量 SECRETID 表示用户的 SecretId，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretID: os.Getenv("TC_SECRET_ID"), // 用户的 SecretId，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参见 https://cloud.tencent.com/document/product/598/37140
			// 环境变量 SECRETKEY 表示用户的 SecretKey，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretKey: os.Getenv("TC_SECRET_KEY"), // 用户的 SecretKey，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参见 https://cloud.tencent.com/document/product/598/37140
		},
	})
}

func TCCosUpload(cosPath, cosName, filePath string) error {
	if cli == nil {
		return errors.New("Tencent COS client not init")
	}
	key := fmt.Sprintf("%s/%s", cosPath, cosName)
	if len(cosPath) == 0 {
		key = cosName
	}

	_, _, err := cli.Object.Upload(
		context.Background(), key, filePath, nil,
	)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "upload file to tencent cos",
			"error":  err,
		}).Send()
	}
	return err
}

func TCCosDownload(cosPath, cosName, filePath string) error {
	if cli == nil {
		return errors.New("Tencent COS client not init")
	}

	key := fmt.Sprintf("%s/%s", cosPath, cosName)
	if len(cosPath) == 0 {
		key = cosName
	}

	opt := &cos.MultiDownloadOptions{
		ThreadPoolSize: 5,
	}
	_, err := cli.Object.Download(
		context.Background(), key, filePath, opt,
	)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "download file from tencent cos",
			"error":  err,
		}).Send()
	}
	return err
}
