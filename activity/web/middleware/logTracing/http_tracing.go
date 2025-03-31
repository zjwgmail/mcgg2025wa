package logTracing

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"net/http"
)

func BuildHttpTracing(ctx context.Context, req *http.Request, tags opentracing.Tags) (opentracing.Span, context.Context) {
	//if config.ApplicationConfig.Opentracing.Enable {
	//	tracer := opentracing.GlobalTracer()
	//	// 生成一个请求的span
	//	clientSpan, clientCtx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("HTTP %v: %v", req.Method, req.URL.Path))
	//	for k, v := range tags {
	//		clientSpan.SetTag(k, v)
	//	}
	//	carrier := opentracing.HTTPHeadersCarrier{}
	//	tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, carrier)
	//	// 将clientSpan的trace-id传递到http header中
	//	for key, value := range carrier {
	//		req.Header.Add(key, value[0])
	//	}
	//	return clientSpan, clientCtx
	//}
	return nil, ctx
}
