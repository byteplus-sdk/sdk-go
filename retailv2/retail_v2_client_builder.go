package retailv2

import (
	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/metrics"
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

func (receiver *ClientBuilder) HostHeader(hostHeader string) *ClientBuilder {
	receiver.param.HostHeader = hostHeader
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

func (receiver *ClientBuilder) MetricsConfig(metricsConfig *metrics.Config) *ClientBuilder {
	receiver.param.MetricsConfig = metricsConfig
	return receiver
}

func (receiver *ClientBuilder) HostAvailablerConfig(hostAvailablerConfig *core.HostAvailablerConfig) *ClientBuilder {
	receiver.param.HostAvailablerConfig = hostAvailablerConfig
	return receiver
}

func (receiver *ClientBuilder) Build() (Client, error) {
	receiver.param.UseAirAuth = true
	context, err := core.NewContext(&receiver.param)
	if err != nil {
		return nil, err
	}
	ru := receiver.buildRetailURL(context)
	httpCaller := core.NewHTTPCaller(context)
	hostAvailabler := core.NewHostAvailabler(ru, context)
	metrics.Collector.Init(context.MetricsConfig(), hostAvailabler)
	client := &clientImpl{
		Client:  common.NewClient(httpCaller, ru.cu),
		hCaller: httpCaller,
		ru:      ru,
		hostAva: hostAvailabler,
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
