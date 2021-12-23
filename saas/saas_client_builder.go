package saas

import (
	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core"
)

type ClientBuilder struct {
	param core.ContextParam
}

func (receiver *ClientBuilder) AK(ak string) *ClientBuilder {
	receiver.param.AK = ak
	return receiver
}

func (receiver *ClientBuilder) SK(sk string) *ClientBuilder {
	receiver.param.SK = sk
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

const saasTenant = "saas"

func (receiver *ClientBuilder) Build() (Client, error) {
	receiver.param.Tenant = saasTenant
	context, err := core.NewContext(&receiver.param)
	if err != nil {
		return nil, err
	}
	su := receiver.buildSaasURL(context)
	httpCaller := core.NewHttpCaller(context)
	client := &clientImpl{
		Client:  common.NewClient(httpCaller, su.su),
		hCaller: httpCaller,
		su:      su,
		hostAva: core.NewHostAvailabler(su, context),
	}
	return client, nil
}

func (receiver *ClientBuilder) buildSaasURL(context *core.Context) *saasURL {
	su := &saasURL{
		schema:    context.Schema(),
		projectId: context.Tenant(),
		su:        common.NewURL(context),
	}
	su.Refresh(context.Hosts()[0])
	return su
}
