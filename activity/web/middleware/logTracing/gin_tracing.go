package logTracing

import (
	"github.com/gin-gonic/gin"
)

var GinTracing = func(ctx *gin.Context) {
	//if config.ApplicationConfig.Opentracing.Enable {
	//	var sp opentracing.Span
	//	opName := "WEB request:" + ctx.Request.URL.Path
	//	// Attempt to join a trace by getting trace context from the headers.
	//	wireContext, err := opentracing.GlobalTracer().Extract(
	//		opentracing.HTTPHeaders,
	//		opentracing.HTTPHeadersCarrier(ctx.Request.Header))
	//	if err != nil {
	//		// If for whatever reason we can't join, go ahead an start a new root span.
	//		sp = opentracing.StartSpan(opName)
	//	} else {
	//		sp = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
	//	}
	//	defer sp.Finish()
	//	ctx.Set("traceSpan", sp)
	//	ctx.Next()
	//}
	ctx.Next()
}
