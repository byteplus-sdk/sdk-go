package core

import "errors"

type ContextParam struct {
	Tenant   string
	TenantId string
	Token    string
	Schema   string
	Hosts    []string
	Headers  map[string]string
	Region   Region
}

func (receiver *ContextParam) checkRequiredField(param *ContextParam) error {
	if param.Tenant == "" {
		return errors.New("tenant is null")
	}
	if param.TenantId == "" {
		return errors.New("tenant id is null")
	}
	if param.Token == "" {
		return errors.New("token is null")
	}
	if param.Region == RegionUnknown {
		return errors.New("region is null")
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
		hosts:           param.Hosts,
		customerHeaders: param.Headers,
	}
	result.fillHosts(param)
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

	// Schema of URL, server supports both "HTTPS" and "HTTP",
	// in order to ensure communication security, please use "HTTPS"
	schema string

	// Server address, china use "rec-b.volcengineapi.com",
	// other area use "tob.sgsnssdk.com" in default
	hosts []string

	// Customer-defined http headers, all requests will include these headers
	customerHeaders map[string]string
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

func (receiver *Context) Schema() string {
	return receiver.schema
}

func (receiver *Context) Hosts() []string {
	return receiver.hosts
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
	receiver.hosts = sgHosts
}

func (receiver *Context) fillDefault() {
	if receiver.schema == "" {
		receiver.schema = "https"
	}
}
