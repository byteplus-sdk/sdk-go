package core

import (
	"errors"

	"github.com/valyala/fasthttp"
)

type ContextParam struct {
	Tenant     string
	TenantId   string
	Token      string
	AK         string
	SK         string
	Schema     string
	HostHeader string
	Hosts      []string
	Headers    map[string]string
	Region     Region
	UseAirAuth bool
}

func (receiver *ContextParam) checkRequiredField(param *ContextParam) error {
	if param.Tenant == "" {
		return errors.New("tenant is null")
	}
	if param.TenantId == "" {
		return errors.New("tenant id is null")
	}
	if err := receiver.checkAuthRequiredField(param); err != nil {
		return err
	}
	if param.Region == RegionUnknown {
		return errors.New("region is null")
	}
	return nil
}

func (receiver *ContextParam) checkAuthRequiredField(param *ContextParam) error {
	if param.UseAirAuth && param.Token == "" {
		return errors.New("token is null")
	}

	if !param.UseAirAuth && (param.AK == "" || param.SK == "") {
		return errors.New("ak or sk is null")
	}

	return nil
}

func NewContext(param *ContextParam) (*Context, error) {
	err := param.checkRequiredField(param)
	if err != nil {
		return nil, err
	}
	result := &Context{
		tenant:          param.Tenant,
		tenantId:        param.TenantId,
		token:           param.Token,
		schema:          param.Schema,
		hostHeader:      param.HostHeader,
		hosts:           param.Hosts,
		customerHeaders: param.Headers,
		useAirAuth:      param.UseAirAuth,
	}
	result.fillHosts(param)
	result.fillVolcCredentials(param)
	if param.HostHeader != "" {
		result.hostHTTPCli = &fasthttp.HostClient{Addr: result.hosts[0]}
	} else {
		result.defaultHTTPCli = &fasthttp.Client{}
	}
	result.fillDefault()
	return result, nil
}

type Context struct {
	// A unique token assigned by bytedance, which is used to
	// generate an authenticated signature when building a request.
	// It is sometimes called "secret".
	tenant string

	// A unique token assigned by bytedance, which is used to
	// generate an authenticated signature when building a request.
	// It is sometimes called "secret".
	tenantId string

	// A unique identity assigned by Bytedance, which is need to fill in URL.
	// It is sometimes called "company".
	token string

	volcCredentials Credential

	// Schema of URL, server supports both "HTTPS" and "HTTP",
	// in order to ensure communication security, please use "HTTPS"
	schema string

	// This field will be used when use ips to request server,
	// it will be set in http header, which named "Host"
	hostHeader string

	// Server address, china use "rec-b.volcengineapi.com",
	// other area use "tob.sgsnssdk.com" in default
	hosts []string

	// Customer-defined http headers, all requests will include these headers
	customerHeaders map[string]string

	// fasthttp default client not support define host
	hostHTTPCli *fasthttp.HostClient

	defaultHTTPCli *fasthttp.Client

	// use air auth, otherwise use volc auth
	useAirAuth bool
}

func (receiver *Context) Tenant() string {
	return receiver.tenant
}

func (receiver *Context) TenantId() string {
	return receiver.tenantId
}

func (receiver *Context) Token() string {
	return receiver.token
}

func (receiver *Context) AK() string {
	return receiver.volcCredentials.AccessKeyID
}

func (receiver *Context) SK() string {
	return receiver.volcCredentials.SecretAccessKey
}

func (receiver *Context) Schema() string {
	return receiver.schema
}

func (receiver *Context) HostHeader() string {
	return receiver.hostHeader
}

func (receiver *Context) Hosts() []string {
	return receiver.hosts
}

func (receiver *Context) UseAirAuth() bool {
	return receiver.useAirAuth
}

func (receiver *Context) UseVolcAuth() bool {
	return !receiver.useAirAuth
}

func (receiver *Context) CustomerHeaders() map[string]string {
	return receiver.customerHeaders
}

func (receiver *Context) fillHosts(param *ContextParam) {
	if len(param.Hosts) > 0 {
		receiver.hosts = param.Hosts
		return
	}
	if param.Region == RegionCn {
		receiver.hosts = cnHosts
		return
	}
	if param.Region == RegionSg {
		receiver.hosts = sgHosts
		return
	}
	if param.Region == RegionUs {
		receiver.hosts = usHosts
		return
	}
	if param.Region == RegionAirCn {
		receiver.hosts = airCnHosts
		return
	}
	if param.Region == RegionAirSg {
		receiver.hosts = airSgHosts
		return
	}
	if param.Region == RegionSaasSg {
		receiver.hosts = saasSgHosts
		return
	}
}

func (receiver *Context) fillDefault() {
	if receiver.schema == "" {
		receiver.schema = "https"
	}
}

func (receiver *Context) fillVolcCredentials(param *ContextParam) {
	c := Credential{
		AccessKeyID:     param.AK,
		SecretAccessKey: param.SK,
		Service:         volcAuthService,
	}

	// fill region
	switch param.Region {
	case RegionSg, RegionAirSg, RegionSaasSg:
		c.Region = "ap-singapore-1"
	case RegionUs:
		c.Region = "us-east-1"
	default: //Region "CN" and "AIR_CN" belong to "cn-north-1"
		c.Region = "cn-north-1"
	}

	receiver.volcCredentials = c
}
