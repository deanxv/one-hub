package transformer

import (
	"fmt"
	"net/http"
	"one-api/providers/claude"
	"one-api/types"
)

// TransformManager 转换管理器
type TransformManager struct {
	sourceTransformer Transformer
	targetTransformer Transformer
}

// NewTransformManager 创建转换管理器
func NewTransformManager(sourceTransformer, targetTransformer Transformer) *TransformManager {
	return &TransformManager{
		sourceTransformer: sourceTransformer,
		targetTransformer: targetTransformer,
	}
}

// ProcessRequest handles request transformation
// Claude request → unified format → VertexGemini format
func (tm *TransformManager) ProcessRequest(claudeRequest *claude.ClaudeRequest) (interface{}, error) {
	// step 1: Claude request to unified format
	unified, err := tm.sourceTransformer.TransformRequestOut(claudeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to transform Claude request to unified format: %v", err)
	}

	// step 2: unified format to VertexGemini format
	targetRequest, err := tm.targetTransformer.TransformRequestIn(unified)
	if err != nil {
		return nil, fmt.Errorf("failed to transform unified format to target format: %v", err)
	}

	return targetRequest, nil
}

// ProcessResponse handles response transformation
// VertexGemini response → unified format → Claude format
func (tm *TransformManager) ProcessResponse(response *http.Response) (interface{}, error) {
	// step 1: VertexGemini response to unified format
	unified, err := tm.targetTransformer.TransformResponseOut(response)
	if err != nil {
		return nil, fmt.Errorf("failed to transform target response to unified format: %v", err)
	}

	// step 2: unified format to Claude format
	claudeResponse, err := tm.sourceTransformer.TransformResponseIn(unified)
	if err != nil {
		return nil, fmt.Errorf("failed to transform unified format to Claude format: %v", err)
	}

	return claudeResponse, nil
}

// ProcessStreamResponse handles stream response transformation
// VertexGemini stream response → unified format stream → Claude stream format
func (tm *TransformManager) ProcessStreamResponse(response *http.Response) (*http.Response, error) {
	// step 1: VertexGemini stream response to unified format stream
	unifiedStream, err := tm.targetTransformer.TransformStreamResponseOut(response)
	if err != nil {
		return nil, fmt.Errorf("failed to transform target stream to unified format: %v", err)
	}

	// step 2: unified format stream to Claude stream format
	claudeStream, err := tm.sourceTransformer.TransformStreamResponseIn(unifiedStream)
	if err != nil {
		return nil, fmt.Errorf("failed to transform unified stream to Claude format: %v", err)
	}

	return claudeStream, nil
}

// UpdateUsage updates provider usage statistics
func (tm *TransformManager) UpdateUsage(provider interface{}, unified *UnifiedChatResponse) {
	if unified.Usage == nil {
		return
	}

	// try to get provider's GetUsage method
	type UsageProvider interface {
		GetUsage() *types.Usage
	}

	if usageProvider, ok := provider.(UsageProvider); ok {
		if usage := usageProvider.GetUsage(); usage != nil {
			usage.PromptTokens = unified.Usage.PromptTokens
			usage.CompletionTokens = unified.Usage.CompletionTokens
			usage.TotalTokens = unified.Usage.TotalTokens
		}
	}
}

// CreateClaudeToVertexGeminiManager 创建 Claude 到 VertexGemini 的转换管理器
func CreateClaudeToVertexGeminiManager() *TransformManager {
	claudeTransformer := NewClaudeTransformer()
	vertexGeminiTransformer := NewVertexGeminiTransformer()
	return NewTransformManager(claudeTransformer, vertexGeminiTransformer)
}
