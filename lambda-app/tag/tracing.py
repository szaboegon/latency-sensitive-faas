from opentelemetry import trace
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.baggage.propagation import W3CBaggagePropagator
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

service_name = "tag"

def instrument_app():
    trace.set_tracer_provider(
    TracerProvider(
        resource=Resource.create({SERVICE_NAME: service_name})
    )
    )
    otel_exporter = OTLPSpanExporter(endpoint="http://otel-collector.observability:9411/api/v2/spans", insecure=True)
    # jaeger_exporter = JaegerExporter(
    # collector_endpoint="http://jaeger-collector.observability.svc.cluster.local:14268/api/traces"
    # )
    trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(otel_exporter)
    )
    RequestsInstrumentor().instrument(tracer_provider=trace.get_tracer_provider())
    tracer = trace.get_tracer(__name__)
    return tracer


