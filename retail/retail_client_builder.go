package retail

import "github.com/byteplus-sdk/sdk-go/core"

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
	ru := &retailURL{
		context: context,
	}
	client := &clientImpl{
		httpCli: core.NewHttpCaller(context),
		ru:      ru,
	}
	initRetailURL(client.ru)
	return client, nil
}

func initRetailURL(ru *retailURL) {
	context := ru.context
	ru.Refresh(context.Hosts()[0])
	ru.hostAvailabler = core.NewHostAvailabler(ru, context)
}
