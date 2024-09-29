from opentelemetry import trace
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.baggage.propagation import W3CBaggagePropagator
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
from opentelemetry.instrumentation.requests import RequestsInstrumentor

service_name = "cut"

def instrument_app():
    trace.set_tracer_provider(
    TracerProvider(
        resource=Resource.create({SERVICE_NAME: service_name})
    )
    )
    jaeger_exporter = JaegerExporter(
    collector_endpoint="http://jaeger-collector.observability.svc.cluster.local:14268/api/traces"
    )

    trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(jaeger_exporter)
    )
    RequestsInstrumentor().instrument(tracer_provider=trace.get_tracer_provider())
    tracer = trace.get_tracer(__name__)
    return tracer


