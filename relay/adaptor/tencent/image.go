package tencent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

func asyncTask(c *gin.Context, jobID, action string) (*TencentTaskResponse, []byte, error) {
	meta := metalib.GetByContext(c)
	// 构造URL
	var url string
	switch action {
	case "TextToImageLite", "SubmitHunyuanImageJob", "QueryHunyuanImageJob", "SubmitHunyuanImageChatJob", "QueryHunyuanImageChatJob", "SubmitHunyuanTo3DJob", "QueryHunyuanTo3DJob":
		url = "https://hunyuan.tencentcloudapi.com/"
	case "SubmitTrainPortraitModelJob", "QueryTrainPortraitModelJob", "SubmitMemeJob", "QueryMemeJob", "SubmitGlamPicJob", "QueryGlamPicJob", "ChangeClothes", "ReplaceBackground", "SketchToImage", "RefineImage", "ImageInpaintingRemoval", "ImageOutpainting", "ImageToImage", "GenerateAvatar":
		url = "https://aiart.tencentcloudapi.com/"
	default:
		return nil, nil, fmt.Errorf("unsupported action: %s", action)
	}
	//构造请求体
	requestBodyMap := map[string]any{
		"JobId": jobID,
		// 一个特例我懒得适配了 : ModelId，仅用于 QueryTrainPortraitModelJob 查询训练写真模型任务
		//这个特例是总结"异步的 分成提交 查询 任务两个步骤的 总结一个通用的 查询 任务 的 请求结构体"时发现遗漏的
	}
	requestBody, err := json.Marshal(requestBodyMap)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, nil, err
	}
	// 构造请求头
	// adaptor.SetupCommonRequestHeader(c, req, meta)                 // 通用请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	action = metalib.GetMappedModelName(action, TaskActionMapping) //查询action->任务action
	meta.VendorContext["Action"] = action
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Region", "ap-guangzhou")
	switch action {
	case "SubmitTrainPortraitModelJob", "QueryTrainPortraitModelJob", "SubmitMemeJob", "QueryMemeJob", "SubmitGlamPicJob", "QueryGlamPicJob", "ChangeClothes", "ReplaceBackground", "SketchToImage", "RefineImage", "ImageInpaintingRemoval", "ImageOutpainting", "ImageToImage", "GenerateAvatar", "UploadTrainPortraitImages", "SubmitDrawPortraitJob", "QueryDrawPortraitJob":
		req.Header.Set("X-TC-Version", "2022-12-29")
		meta.VendorContext["Version"] = "2022-12-29"
	case "SubmitHunyuanImageJob", "QueryHunyuanImageJob", "SubmitHunyuanImageChatJob", "QueryHunyuanImageChatJob", "TextToImageLite", "SubmitHunyuanTo3DJob", "QueryHunyuanTo3DJob":
		req.Header.Set("X-TC-Version", "2023-09-01")
		meta.VendorContext["Version"] = "2023-09-01"
	default:
		return nil, nil, fmt.Errorf("unsupported action: %s", action)
	}
	sign := GetSign(requestBodyMap, meta) // 使用结构化数据而不是序列化后的requestBody
	req.Header.Set("Authorization", sign)
	req.Header.Set("X-TC-Timestamp", meta.VendorContext["Timestamp"].(string)) //在GetSign里设置过meta.VendorContext["Timestamp"]
	// req.Header.Set("X-TC-Nonce", strconv.FormatInt(time.Now().UnixNano(), 10))
	client := &http.Client{} // 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body) // 读取响应
	if err != nil {
		return nil, nil, err
	}
	var tencentResponseP struct { // 解析响应
		Response TencentTaskResponse `json:"Response"`
	}
	err = json.Unmarshal(responseBody, &tencentResponseP)
	if err != nil {
		return nil, responseBody, err
	}
	if tencentResponseP.Response.Error != nil && tencentResponseP.Response.Error.Code != "" {
		return &tencentResponseP.Response, responseBody, fmt.Errorf("API error: %s - %s",
			tencentResponseP.Response.Error.Code,
			tencentResponseP.Response.Error.Message)
	}
	return &tencentResponseP.Response, responseBody, nil
}

// asyncTaskWait 轮询等待任务完成
func asyncTaskWait(c *gin.Context, jobID, action string) (*TencentTaskResponse, error) {
	waitSeconds := 1
	maxStep := 20
	var taskResponse *TencentTaskResponse
	for step := 0; step < maxStep; step++ {
		response, _, err := asyncTask(c, jobID, action)
		if err != nil {
			return nil, err
		}
		taskResponse = response
		// 检查任务状态
		if action == "QueryHunyuanTo3DJob" {
			// 3D模型生成特殊处理
			if taskResponse.Status == "DONE" {
				return taskResponse, nil
			} else if taskResponse.Status == "FAILED" || taskResponse.Status == "CANCELED" {
				return nil, fmt.Errorf("task_failed: %s", taskResponse.Status)
			}
			// 如果是 INIT/WAIT/RUN 状态则什么也不做 直接去执行sleep
		} else {
			// 根据状态码判断
			switch taskResponse.JobStatusCode {
			case "5", "DONE": // 处理完成
				return taskResponse, nil
			case "4", "FAIL": // 处理失败
				var errMsg string
				if taskResponse.JobErrorMsg != "" {
					errMsg = taskResponse.JobErrorMsg
				} else {
					errMsg = taskResponse.JobStatusMsg
				}
				return nil, fmt.Errorf("task_failed: %s,job error code: %s", errMsg, taskResponse.JobErrorCode)
			case "1", "2", "INIT", "WAIT", "RUN": //1：等待中、2：运行中 INIT: 初始化、WAIT：等待中、RUN：运行中
				// 继续循环
			default: // 处理其他未知状态码
				var message string
				if taskResponse.JobErrorMsg != "" {
					message = taskResponse.JobErrorMsg
				} else {
					message = taskResponse.JobStatusMsg
				}
				return nil, fmt.Errorf("unknown_status: %s, message: %s", taskResponse.JobStatusCode, message)
			}
			// 如果有错误码也认为是失败
			if taskResponse.JobErrorCode != "" {
				return nil, fmt.Errorf("task_failed: %s - %s", taskResponse.JobErrorCode, taskResponse.JobErrorMsg)
			}
		}
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}
	return nil, fmt.Errorf("async_task_wait_timeout: job_id=%s", jobID)
}

func ImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	meta := metalib.GetByContext(c)
	var tencentResponse TencentTaskResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var tencentResponseP struct {
		Response TencentTaskResponse `json:"Response,omitempty"`
	}
	err = json.Unmarshal(responseBody, &tencentResponseP)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	tencentResponse = tencentResponseP.Response
	// 检查上一步 提交任务 是否出错
	if tencentResponse.Error != nil && tencentResponse.Error.Code != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: tencentResponse.Error.Message,
				Code:    tencentResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	action := meta.VendorContext["Action"].(string)
	switch action {
	// 异步任务，需要轮询
	case "SubmitTrainPortraitModelJob", "SubmitHunyuanImageJob", "SubmitHunyuanImageChatJob", "SubmitHunyuanTo3DJob", "SubmitMemeJob", "SubmitGlamPicJob", "QueryTrainPortraitModelJob", "QueryHunyuanImageJob", "QueryHunyuanImageChatJob", "QueryHunyuanTo3DJob", "QueryMemeJob", "QueryGlamPicJob":
		finalResponse, err := asyncTaskWait(c, tencentResponse.JobId, action)
		if err != nil {
			return openai.ErrorWrapper(err, "async_task_wait_failed", http.StatusInternalServerError), nil
		}
		// 使用异步任务的最终响应替换原响应
		tencentResponse = *finalResponse
	// 同步任务 直接返回
	case "TextToImageLite", "ImageToImage", "GenerateAvatar", "ChangeClothes", "ReplaceBackground", "SketchToImage", "RefineImage", "ImageInpaintingRemoval", "ImageOutpainting":
		//文生图轻量版 接口：TextToImageLite 描述：根据文本描述直接生成图像，返回结果图URL或Base64编码。图像风格化（图生图） 接口：ImageToImage 描述：输入图像和文本描述，直接返回风格化后的图像。百变头像  接口：GenerateAvatar  描述：根据输入的人像照片生成风格化头像，直接返回结果。模特换装  接口：ChangeClothes  描述：输入模特图和服装图，直接返回换装后的图像。商品背景生成  接口：ReplaceBackground=  描述：替换商品图的背景，直接返回结果。线稿生图  接口：SketchToImage  描述：输入线稿图和文本描述，直接返回上色后的图像。图片变清晰  接口：RefineImage  描述：增强图像清晰度，直接返回结果。局部消除  接口：ImageInpaintingRemoval  描述：根据Mask消除指定区域，直接返回修复后的图像。扩图  接口：ImageOutpainting  描述：按比例扩展图像边缘，直接返回结果。
	}

	// 任务成功完成，构建 OpenAI 格式的响应
	fullTextResponse := responseTencent2OpenAIImage(&tencentResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "write_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, nil
}

// responseTencent2OpenAIImage 将腾讯混元的响应转换为 OpenAI 格式的图像响应
func responseTencent2OpenAIImage(response *TencentTaskResponse) *openai.ImageResponse {
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
	}
	// 处理单个图像URL或图像URL数组
	var imageUrls []string
	switch v := response.ResultImage.(type) {
	case string:
		imageUrls = []string{v}
	case []string:
		imageUrls = v
	case []interface{}:
		imageUrls = make([]string, len(v))
		for i, url := range v {
			if strUrl, ok := url.(string); ok {
				imageUrls[i] = strUrl
			}
		}
	}
	// 将所有图像URL添加到响应中
	for i, url := range imageUrls {
		// 可以根据需要处理修改后的提示词
		var revisedPrompt string
		if response.RevisedPrompt != nil && len(response.RevisedPrompt) > i {
			revisedPrompt = response.RevisedPrompt[i]
		}
		imageResponse.Data = append(imageResponse.Data, openai.ImageData{
			Url:           url,
			B64Json:       "", // RspImgType似乎可以控制返回类型 但暂时不写了 烦
			RevisedPrompt: revisedPrompt,
		})
	}
	// 将 3D 文件URL也当作普通图像URL返回(暂时，未来应该会单独开/v1/3d/generations)
	if response.ResultFile3Ds != nil {
		for _, file3Ds := range response.ResultFile3Ds {
			for _, file3D := range file3Ds.File3D {
				imageResponse.Data = append(imageResponse.Data, openai.ImageData{
					Url:           file3D.Url,
					B64Json:       "",
					RevisedPrompt: "", // 3D 文件通常不需要 revisedPrompt，可按需设置
				})
			}
		}
	}
	return &imageResponse
}
