package retail

import (
	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core"
)

type ClientBuilder struct {
	param core.ContextParam
}

func (receiver *ClientBuilder) Tenant(tenant string) *ClientBuilder {
	receiver.param.Tenant = tenant
	return receiver
}

func (receiver *ClientBuilder) TenantId(tenantId string) *ClientBuilder {
	receiver.param.TenantId = tenantId
	return receiver
}

func (receiver *ClientBuilder) Token(token string) *ClientBuilder {
	receiver.param.Token = token
	return receiver
}

func (receiver *ClientBuilder) Schema(schema string) *ClientBuilder {
	receiver.param.Schema = schema
	return receiver
}

func (receiver *ClientBuilder) HostHeader(host string) *ClientBuilder {
	receiver.param.HostHeader = host
	return receiver
}

func (receiver *ClientBuilder) Hosts(hosts []string) *ClientBuilder {
	receiver.param.Hosts = hosts
	return receiver
}

func (receiver *ClientBuilder) Headers(headers map[string]string) *ClientBuilder {
	receiver.param.Headers = headers
	return receiver
}

func (receiver *ClientBuilder) Region(region core.Region) *ClientBuilder {
	receiver.param.Region = region
	return receiver
}

func (receiver *ClientBuilder) Build() (Client, error) {
	context, err := core.NewContext(&receiver.param)
	if err != nil {
		return nil, err
	}
	ru := receiver.buildRetailURL(context)
	httpCaller := core.NewHttpCaller(context)
	client := &clientImpl{
		cCli:    common.NewClient(httpCaller, ru.cu),
		hCaller: httpCaller,
		ru:      ru,
		hostAva: core.NewHostAvailabler(ru, context),
	}
	return client, nil
}

func (receiver *ClientBuilder) buildRetailURL(context *core.Context) *retailURL {
	ru := &retailURL{
		schema: context.Schema(),
		tenant: context.Tenant(),
		cu:     common.NewURL(context),
	}
	ru.Refresh(context.Hosts()[0])
	return ru
}
