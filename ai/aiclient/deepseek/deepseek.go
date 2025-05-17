package deepseek

import (
	"context"
	"lygPP/mk_agent/ai/httputil"
	"lygPP/mk_agent/model"
)

type DeepSeek struct {
	Cfg httputil.Config
}

func NewDeepSeek(cfg httputil.Config) *DeepSeek {
	return &DeepSeek{Cfg: cfg}
}

func (d *DeepSeek) Stream(ctx context.Context, request model.Request) (<-chan model.Response, error) {
	return httputil.HandleStreaming(ctx, request, &d.Cfg)
}

func (d *DeepSeek) Chat(ctx context.Context, request model.Request) (*model.Response, error) {
	doRequest, err := httputil.DoRequest(ctx, request, false, &d.Cfg)
	if err != nil {
		return nil, err
	}
	return httputil.ParseResponse(doRequest)
}
