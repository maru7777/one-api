package cloudflare

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// ImageHandler 处理 Cloudflare Workers AI 生成的图像响应
func ImageHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	var imageData []byte
	meta := meta.GetByContext(c)
	// 1. 根据模型类型解析响应
	if modelName == "@cf/black-forest-labs/flux-1-schnell" {
		// flux-1-schnell 返回 JSON 格式
		var imageResp CloudflareImageResponse
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
		}
		err = json.Unmarshal(responseBody, &imageResp)
		if err != nil {
			return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
		}
		if len(imageResp.Errors) > 0 {
			var errMsg string = fmt.Sprintf("Code: %d, Message: %s", imageResp.Errors[0].Code, imageResp.Errors[0].Message)
			return openai.ErrorWrapper(errors.New(errMsg), "cloudflare_image_generation_failed", http.StatusInternalServerError), nil
		}
		// 解码 Base64 数据
		if len(imageResp.Result) > 0 {
			data, err := base64.StdEncoding.DecodeString(imageResp.Result["image"])
			if err != nil {
				logger.SysError("error decoding base64 image: " + err.Error())
				return openai.ErrorWrapper(err, "decode_base64_failed", http.StatusInternalServerError), nil
			}
			imageData = data
		}
	} else {
		// 其他模型返回二进制 PNG 数据，直接读取
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.SysError("error reading response body: " + err.Error())
			return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
		}
		imageData = data
	}

	// 2. 关闭响应体
	err := resp.Body.Close()
	if err != nil {
		logger.SysError("error closing response body: " + err.Error())
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// 3. 上传图像到 Cloudflare Images
	imageURL, uploadErr := uploadToCloudflareR2(meta, imageData)
	if uploadErr != nil {
		logger.SysError("error uploading image to Cloudflare: " + uploadErr.Error())
		return openai.ErrorWrapper(uploadErr, "upload_image_failed", http.StatusInternalServerError), nil
	}

	// 4. 构造 OpenAI 兼容格式的响应，使用已有的结构
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
		Data: []openai.ImageData{
			{
				Url: imageURL,
			},
		},
	}

	// 设置响应头并返回 JSON
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	jsonResponse, _ := json.Marshal(imageResponse)
	_, _ = c.Writer.Write(jsonResponse)

	// 5. 返回 Usage（图像生成不需要详细的 token 计数）
	usage := &model.Usage{
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
	}
	return nil, usage
}

// uploadToCloudflareR2 上传图像到 Cloudflare R2 存储桶
func uploadToCloudflareR2(meta *meta.Meta, imageData []byte) (string, error) {
	// 获取 R2 配置
	accessKeyID := os.Getenv("CLOUDFLARE_R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("CLOUDFLARE_R2_SECRET_ACCESS_KEY")
	accountID := meta.Config.UserID
	bucketName := "test"
	publicBucketURL := "https://pub-ff7617225b80468fa7510ad087fd831a.r2.dev"
	// 验证配置是否完整
	if accessKeyID == "" || secretAccessKey == "" || accountID == "" || bucketName == "" || publicBucketURL == "" {
		return "", fmt.Errorf("missing Cloudflare R2 configuration")
	}

	// 构造 R2 端点 URL（用于上传）
	endpointURL := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	// 创建 AWS 会话和客户端
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID, secretAccessKey, "",
		)),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: endpointURL,
				}, nil
			},
		)),
	)
	if err != nil {
		return "", fmt.Errorf("create AWS config failed: %v", err)
	}

	// 创建 S3 客户端
	s3Client := s3.NewFromConfig(cfg)

	// 生成唯一的对象键（文件名）
	objectKey := fmt.Sprintf("images/%s.png", random.GetUUID())

	// 上传文件到 R2，设置 ACL 为 public-read
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String("image/png"),
		ACL:         types.ObjectCannedACLPublicRead, // 使用枚举值
	})
	if err != nil {
		logger.SysError("error uploading to R2: " + err.Error())
		return "", fmt.Errorf("upload to R2 failed: %v", err)
	}

	// 返回公共 URL
	imageURL := fmt.Sprintf("%s/%s", publicBucketURL, objectKey)

	return imageURL, nil
}
