package common

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const KeyRequestBody = "key_request_body"

func GetRequestBody(c *gin.Context) ([]byte, error) {
	requestBody, _ := c.Get(KeyRequestBody)
	if requestBody != nil {
		return requestBody.([]byte), nil
	}
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read request body failed")
	}
	_ = c.Request.Body.Close()
	c.Set(KeyRequestBody, requestBody)
	return requestBody.([]byte), nil
}

func UnmarshalBodyReusable(c *gin.Context, v any) error {
	requestBody, err := GetRequestBody(c)
	if err != nil {
		return errors.Wrap(err, "get request body failed")
	}

	// Reset request body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	defer func() {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}()

	if err = c.Bind(v); err != nil {
		return errors.Wrap(err, "bind request body failed")
	}

	return nil
}

func SetEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}
