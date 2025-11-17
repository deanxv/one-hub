package gemini

import (
	"io"
	"net/http"
	"one-api/common"
	"one-api/types"
	"strings"
	"time"
)

// CreateVeoVideoAndDownload 提交 Veo 3.0 视频生成任务，轮询直到完成，然后下载视频文件
// 重要：整个流程使用同一个 Provider 实例，确保使用相同的 API Key
func (p *GeminiProvider) CreateVeoVideoAndDownload(request *VeoVideoRequest, modelName string) ([]byte, string, *types.OpenAIErrorWithStatusCode) {
	// 1. 提交视频生成任务（使用当前 Provider 的 API Key）
	operation, err := p.submitVeoTask(request, modelName)
	if err != nil {
		return nil, "", err
	}

	// 2. 轮询状态直到完成（使用相同的 Provider 实例）
	finalResponse, err := p.pollVeoStatusUntilDone(operation.Name)
	if err != nil {
		return nil, "", err
	}

	// 3. 提取视频 URI
	if finalResponse.Response == nil ||
		finalResponse.Response.GenerateVideoResponse == nil ||
		len(finalResponse.Response.GenerateVideoResponse.GeneratedSamples) == 0 ||
		finalResponse.Response.GenerateVideoResponse.GeneratedSamples[0].Video == nil {
		return nil, "", common.StringErrorWrapper("no video generated", "no_video", http.StatusInternalServerError)
	}

	videoURI := finalResponse.Response.GenerateVideoResponse.GeneratedSamples[0].Video.Uri
	if videoURI == "" {
		return nil, "", common.StringErrorWrapper("empty video URI", "empty_uri", http.StatusInternalServerError)
	}

	return p.downloadVideoFile(videoURI)
}

// submitVeoTask 提交 Veo 视频生成任务
func (p *GeminiProvider) submitVeoTask(request *VeoVideoRequest, modelName string) (*VeoLongRunningResponse, *types.OpenAIErrorWithStatusCode) {
	// 构建请求 URL，使用传入的模型名称
	fullRequestURL := p.GetFullRequestURL("predictLongRunning", modelName)

	// 获取请求头
	headers := p.GetRequestHeaders()
	headers["Content-Type"] = "application/json"

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(request), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	veoResponse := &VeoLongRunningResponse{}
	// 发送请求
	_, errWithCode := p.Requester.SendRequest(req, veoResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return veoResponse, nil
}

// pollVeoStatusUntilDone 轮询 Veo 状态直到完成
func (p *GeminiProvider) pollVeoStatusUntilDone(operationName string) (*VeoLongRunningResponse, *types.OpenAIErrorWithStatusCode) {
	for {
		response, err := p.GetVeoOperationStatus(operationName)
		if err != nil {
			return nil, err
		}

		// 如果完成，返回结果
		if response.Done {
			return response, nil
		}

		time.Sleep(10 * time.Second)
	}
}

// downloadVideoFile 下载视频文件
// 重要：使用相同的 Provider 实例，确保 API Key 一致
func (p *GeminiProvider) downloadVideoFile(videoURI string) ([]byte, string, *types.OpenAIErrorWithStatusCode) {
	headers := p.GetRequestHeaders()

	req, err := p.Requester.NewRequest(http.MethodGet, videoURI, p.Requester.WithHeader(headers))
	if err != nil {
		return nil, "", common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, "", errWithCode
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", common.ErrorWrapper(err, "read_response_failed", http.StatusInternalServerError)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4" // 默认为 MP4
	}

	return fileData, contentType, nil
}

// GetVeoOperationStatus 查询 Veo 3.0 操作状态
func (p *GeminiProvider) GetVeoOperationStatus(operationName string) (*VeoLongRunningResponse, *types.OpenAIErrorWithStatusCode) {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")
	version := "v1beta"

	if p.Channel.Other != "" {
		version = p.Channel.Other
	}

	inputVersion := p.Context.Param("version")
	if inputVersion != "" {
		version = inputVersion
	}

	fullRequestURL := baseURL + "/" + version + "/" + operationName

	headers := p.GetRequestHeaders()

	req, err := p.Requester.NewRequest(http.MethodGet, fullRequestURL, p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	veoResponse := &VeoLongRunningResponse{}
	_, errWithCode := p.Requester.SendRequest(req, veoResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return veoResponse, nil
}
