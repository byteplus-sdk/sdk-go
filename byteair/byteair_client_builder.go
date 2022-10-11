package byteair

import (
	"github.com/byteplus-sdk/sdk-go/common"
	"github.com/byteplus-sdk/sdk-go/core"
	"github.com/byteplus-sdk/sdk-go/core/metrics"
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

func (receiver *ClientBuilder) AK(ak string) *ClientBuilder {
	receiver.param.AK = ak
	return receiver
}

func (receiver *ClientBuilder) SK(sk string) *ClientBuilder {
	receiver.param.SK = sk
	return receiver
}

func (receiver *ClientBuilder) UseAirAuth() *ClientBuilder {
	receiver.param.UseAirAuth = true
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
	context, err := core.NewContext(&receiver.param)
	if err != nil {
		return nil, err
	}
	gu := receiver.buildByteairURL(context)
	httpCaller := core.NewHTTPCaller(context)
	hostAvailabler := core.NewHostAvailabler(gu, context)
	metrics.Collector.Init(context.MetricsConfig(), hostAvailabler)
	client := &clientImpl{
		Client:  common.NewClient(httpCaller, gu.cu),
		hCaller: httpCaller,
		gu:      gu,
		hostAva: hostAvailabler,
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
