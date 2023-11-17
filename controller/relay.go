package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"strconv"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"
)

type Message struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Name    *string `json:"name,omitempty"`
}

type VisionMessage struct {
	Role    string                     `json:"role"`
	Content OpenaiVisionMessageContent `json:"content"`
	Name    *string                    `json:"name,omitempty"`
}

// OpenaiVisionMessageContentType vision message content type
type OpenaiVisionMessageContentType string

const (
	// OpenaiVisionMessageContentTypeText text
	OpenaiVisionMessageContentTypeText OpenaiVisionMessageContentType = "text"
	// OpenaiVisionMessageContentTypeImageUrl image url
	OpenaiVisionMessageContentTypeImageUrl OpenaiVisionMessageContentType = "image_url"
)

// OpenaiVisionMessageContent vision message content
type OpenaiVisionMessageContent struct {
	Type     OpenaiVisionMessageContentType     `json:"type"`
	Text     string                             `json:"text,omitempty"`
	ImageUrl OpenaiVisionMessageContentImageUrl `json:"image_url,omitempty"`
}

// VisionImageResolution image resolution
type VisionImageResolution string

const (
	// VisionImageResolutionLow low resolution
	VisionImageResolutionLow VisionImageResolution = "low"
	// VisionImageResolutionHigh high resolution
	VisionImageResolutionHigh VisionImageResolution = "high"
)

type OpenaiVisionMessageContentImageUrl struct {
	URL    string                `json:"url"`
	Detail VisionImageResolution `json:"detail,omitempty"`
}

const (
	RelayModeUnknown = iota
	RelayModeChatCompletions
	RelayModeCompletions
	RelayModeEmbeddings
	RelayModeModerations
	RelayModeImagesGenerations
	RelayModeEdits
	RelayModeAudio
)

// https://platform.openai.com/docs/api-reference/chat

type GeneralOpenAIRequest struct {
	Model string `json:"model,omitempty"`
	// Messages maybe []Message or []VisionMessage
	Messages    any     `json:"messages,omitempty"`
	Prompt      any     `json:"prompt,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	N           int     `json:"n,omitempty"`
	Input       any     `json:"input,omitempty"`
	Instruction string  `json:"instruction,omitempty"`
	Size        string  `json:"size,omitempty"`
	Functions   any     `json:"functions,omitempty"`
}

func (r *GeneralOpenAIRequest) MessagesLen() int {
	switch msgs := r.Messages.(type) {
	case []any:
		return len(msgs)
	case []Message:
		return len(msgs)
	case []VisionMessage:
		return len(msgs)
	case []map[string]any:
		return len(msgs)
	default:
		return 0
	}
}

// TextMessages returns messages as []Message
func (r *GeneralOpenAIRequest) TextMessages() (messages []Message, err error) {
	if blob, err := json.Marshal(r.Messages); err != nil {
		return nil, errors.Wrap(err, "marshal messages failed")
	} else if err := json.Unmarshal(blob, &messages); err != nil {
		return nil, errors.Wrap(err, "unmarshal messages failed")
	} else {
		return messages, nil
	}
}

// VisionMessages returns messages as []VisionMessage
func (r *GeneralOpenAIRequest) VisionMessages() (messages []VisionMessage, err error) {
	if blob, err := json.Marshal(r.Messages); err != nil {
		return nil, errors.Wrap(err, "marshal vision messages failed")
	} else if err := json.Unmarshal(blob, &messages); err != nil {
		return nil, errors.Wrap(err, "unmarshal vision messages failed")
	} else {
		return messages, nil
	}
}

func (r GeneralOpenAIRequest) ParseInput() []string {
	if r.Input == nil {
		return nil
	}
	var input []string
	switch r.Input.(type) {
	case string:
		input = []string{r.Input.(string)}
	case []any:
		input = make([]string, 0, len(r.Input.([]any)))
		for _, item := range r.Input.([]any) {
			if str, ok := item.(string); ok {
				input = append(input, str)
			}
		}
	}
	return input
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type TextRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Prompt    string    `json:"prompt"`
	MaxTokens int       `json:"max_tokens"`
	//Stream   bool      `json:"stream"`
}

type ImageRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

type AudioResponse struct {
	Text string `json:"text,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
}

type OpenAIErrorWithStatusCode struct {
	OpenAIError
	StatusCode int `json:"status_code"`
}

type TextResponse struct {
	Choices []OpenAITextResponseChoice `json:"choices"`
	Usage   `json:"usage"`
	Error   OpenAIError `json:"error"`
}

type OpenAITextResponseChoice struct {
	Index        int `json:"index"`
	Message      `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type OpenAITextResponse struct {
	Id      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Choices []OpenAITextResponseChoice `json:"choices"`
	Usage   `json:"usage"`
}

type OpenAIEmbeddingResponseItem struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type OpenAIEmbeddingResponse struct {
	Object string                        `json:"object"`
	Data   []OpenAIEmbeddingResponseItem `json:"data"`
	Model  string                        `json:"model"`
	Usage  `json:"usage"`
}

type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		Url string `json:"url"`
	}
}

type ChatCompletionsStreamResponseChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

type ChatCompletionsStreamResponse struct {
	Id      string                                `json:"id"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Model   string                                `json:"model"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
}

type CompletionsStreamResponse struct {
	Choices []struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func Relay(c *gin.Context) {
	relayMode := RelayModeUnknown
	if strings.HasPrefix(c.Request.URL.Path, "/v1/chat/completions") {
		relayMode = RelayModeChatCompletions
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/completions") {
		relayMode = RelayModeCompletions
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/embeddings") {
		relayMode = RelayModeEmbeddings
	} else if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		relayMode = RelayModeEmbeddings
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		relayMode = RelayModeModerations
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		relayMode = RelayModeImagesGenerations
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/edits") {
		relayMode = RelayModeEdits
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		relayMode = RelayModeAudio
	}
	var err *OpenAIErrorWithStatusCode
	switch relayMode {
	case RelayModeImagesGenerations:
		err = relayImageHelper(c, relayMode)
	case RelayModeAudio:
		err = relayAudioHelper(c, relayMode)
	default:
		err = relayTextHelper(c, relayMode)
	}
	if err != nil {
		requestId := c.GetString(common.RequestIdKey)
		retryTimesStr := c.Query("retry")
		retryTimes, _ := strconv.Atoi(retryTimesStr)
		if retryTimesStr == "" {
			retryTimes = common.RetryTimes
		}
		if retryTimes > 0 {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?retry=%d", c.Request.URL.Path, retryTimes-1))
		} else {
			if err.StatusCode == http.StatusTooManyRequests {
				err.OpenAIError.Message = "当前分组上游负载已饱和，请稍后再试"
			}
			err.OpenAIError.Message = common.MessageWithRequestId(err.OpenAIError.Message, requestId)
			c.JSON(err.StatusCode, gin.H{
				"error": err.OpenAIError,
			})
		}
		channelId := c.GetInt("channel_id")
		common.LogError(c.Request.Context(), fmt.Sprintf("relay error (channel #%d): %s", channelId, err.Message))
		// https://platform.openai.com/docs/guides/error-codes/api-errors
		if shouldDisableChannel(&err.OpenAIError, err.StatusCode) {
			channelId := c.GetInt("channel_id")
			channelName := c.GetString("channel_name")
			disableChannel(channelId, channelName, err.Message)
		}
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := OpenAIError{
		Message: "API not implemented",
		Type:    "one_api_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := OpenAIError{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
