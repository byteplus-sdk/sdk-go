package core

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/byte-plus/sdk-go/core/logs"
	"github.com/byte-plus/sdk-go/core/option"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
	"time"
)

const netErrMark = "[netErr]"

func NewHttpCaller(context *Context) *HttpCaller {
	return &HttpCaller{context: context}
}

type HttpCaller struct {
	context *Context
}

func (c *HttpCaller) DoRequest(url string, request proto.Message,
	response proto.Message, opts ...option.Option) error {
	reqBytes, err := c.marshal(request)
	if err != nil {
		logs.Error("marshal request fail, err:%s url:%s", err.Error(), url)
		return err
	}
	options := option.Conv2Options(opts...)
	headers := c.buildHeaders(options, reqBytes)
	rspBytes, err := c.doHttpRequest(url, headers, reqBytes, options.Timeout)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(rspBytes, response)
	if err != nil {
		logs.Error("unmarshal response fail, err:%s url:%s", err.Error(), url)
		return err
	}
	return nil
}

func (c *HttpCaller) marshal(request proto.Message) ([]byte, error) {
	reqBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	reqBytes = fasthttp.AppendGzipBytes(nil, reqBytes)
	return reqBytes, nil
}

func (c *HttpCaller) buildHeaders(options *option.Options, reqBytes []byte) map[string]string {
	headers := make(map[string]string)
	headers["Content-Encoding"] = "gzip"
	headers["Accept-Encoding"] = "gzip"
	headers["Content-Type"] = "application/x-protobuf"
	headers["Accept"] = "application/x-protobuf"
	if len(options.RequestId) == 0 {
		headers["Request-Id"] = uuid.NewString()
	} else {
		headers["Request-Id"] = options.RequestId
	}
	c.withAuthHeader(headers, reqBytes)
	for k, v := range options.Headers {
		headers[k] = v
	}
	return headers
}

func (c *HttpCaller) withAuthHeader(header map[string]string, reqBytes []byte) {
	var (
		tenantId = c.context.tenantId
		// Gets the second-level timestamp of the current time.
		// The server only supports the second-level timestamp.
		// The 'ts' must be the current time.
		// When current time exceeds a certain time, such as 5 seconds, of 'ts',
		// the signature will be invalid and cannot pass authentication
		ts = strconv.FormatInt(time.Now().Unix(), 10)
		// Use sub string of UUID as "nonce",  too long will be wasted.
		// You can also use 'ts' as' nonce'
		nonce = uuid.NewString()[:8]
		// calculate the authentication signature
		signature = c.calSignature(reqBytes, ts, nonce)
	)

	header["Tenant-Id"] = tenantId
	header["Tenant-Ts"] = ts
	header["Tenant-Nonce"] = nonce
	header["Tenant-Signature"] = signature
}

func (c *HttpCaller) calSignature(reqBytes []byte, ts, nonce string) string {
	var (
		token    = c.context.token
		tenantId = c.context.tenantId
	)
	// Splice in the order of "token", "HttpBody", "tenant_id", "ts", and "nonce".
	// The order must not be mistaken.
	// String need to be encoded as byte arrays by UTF-8
	shaHash := sha256.New()
	shaHash.Write([]byte(token))
	shaHash.Write(reqBytes)
	shaHash.Write([]byte(tenantId))
	shaHash.Write([]byte(ts))
	shaHash.Write([]byte(nonce))
	return fmt.Sprintf("%x", shaHash.Sum(nil))
}

func (c *HttpCaller) doHttpRequest(url string,
	headers map[string]string, reqBytes []byte, timeout time.Duration) ([]byte, error) {
	request := c.acquireRequest(url, headers, reqBytes)
	response := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()

	start := time.Now()
	defer func() {
		logs.Debug("http url:%s, cost:%s", url, time.Now().Sub(start))
	}()
	var err error
	logs.Trace("http request header:\n%s", string(request.Header.Header()))
	if timeout > 0 {
		err = fasthttp.DoTimeout(request, response, timeout)
	} else {
		err = fasthttp.Do(request, response)
	}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			logs.Error("do http request timeout, msg:%s", err.Error())
			return nil, errors.New(netErrMark + " timeout")
		}

		logs.Error("do http request occur error, msg:%s", err.Error())
		return nil, err
	}
	logs.Trace("http response headers:\n%s", string(response.Header.Header()))
	if response.StatusCode() != fasthttp.StatusOK {
		c.logHttpResponse(url, response)
		return nil, errors.New(netErrMark + "http status not 200")
	}
	rspEncoding := response.Header.Peek(fasthttp.HeaderContentEncoding)
	if bytes.Contains(rspEncoding, []byte("gzip")) {
		rspBytes, err := response.BodyGunzip()
		if err != nil {
			rspHeaders := string(response.Header.Header())
			logs.Error("gzip decompress rsp err, url:%s header:\n%s", url, rspHeaders)
			return nil, err
		}
		return rspBytes, nil
	}
	return response.Body(), nil
}

func (c *HttpCaller) acquireRequest(url string,
	headers map[string]string, reqBytes []byte) *fasthttp.Request {
	request := fasthttp.AcquireRequest()
	request.Header.SetMethod(fasthttp.MethodPost)
	request.SetRequestURI(url)
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	request.SetBodyRaw(reqBytes)
	return request
}

func (c *HttpCaller) logHttpResponse(url string, response *fasthttp.Response) {
	rspBytes := response.Body()
	if len(rspBytes) > 0 {
		logs.Error("http status not 200, url:%s code:%d body:\n%s\n",
			url, response.StatusCode(), string(rspBytes))
		return
	}
	logs.Error("http status not 200, url:%s code:%d",
		url, response.StatusCode())
}
