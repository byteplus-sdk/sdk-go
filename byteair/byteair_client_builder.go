package byteair

import (
	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core"
)

type ClientBuilder struct {
	param core.ContextParam
}

func (receiver *ClientBuilder) ProjectId(projectId string) *ClientBuilder {
	receiver.param.Tenant = projectId
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

func (receiver *ClientBuilder) AK(ak string) *ClientBuilder {
	receiver.param.AK = ak
	return receiver
}

func (receiver *ClientBuilder) SK(sk string) *ClientBuilder {
	receiver.param.SK = sk
	return receiver
}

func (receiver *ClientBuilder) Build() (Client, error) {
	context, err := core.NewContext(&receiver.param)
	if err != nil {
		return nil, err
	}
	gu := receiver.buildByteairURL(context)
	httpCaller := core.NewHttpCaller(context)
	client := &clientImpl{
		Client:  common.NewClient(httpCaller, gu.cu),
		hCaller: httpCaller,
		gu:      gu,
		hostAva: core.NewHostAvailabler(gu, context),
	}
	return client, nil
}

func (receiver *ClientBuilder) buildByteairURL(context *core.Context) *byteairURL {
	gu := &byteairURL{
		schema: context.Schema(),
		tenant: context.Tenant(),
		cu:     common.NewURL(context),
	}
	gu.Refresh(context.Hosts()[0])
	return gu
}
