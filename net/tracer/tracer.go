package tracer

import (
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"golang-kit/config"
	"golang-kit/log"
)

func InitTracer(c *config.Trace, cf *config.Common) (tracer opentracing.Tracer, collector zipkin.Collector, err error) {
	if collector, err = zipkin.NewHTTPCollector(c.Addr); err != nil {
		log.Error("zipkin.NewHTTPCollector error(%v)", err)
		return
	}
	recorder := zipkin.NewRecorder(collector, c.Debug, cf.HostPort, cf.Family)
	tracer, err = zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(c.SameSpan),
		zipkin.TraceID128Bit(c.TraceID128Bit),
	)
	if err != nil {
		log.Error("zipkin.NewTracer error(%v)", err)
		return
	}
	opentracing.InitGlobalTracer(tracer)
	return
}
