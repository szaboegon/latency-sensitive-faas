from opentelemetry import trace
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

service_name = "objectdetect"

def instrument_app():
    trace.set_tracer_provider(
        TracerProvider(resource=Resource.create({SERVICE_NAME: service_name}))
    )

    otel_exporter = OTLPSpanExporter(endpoint="otel-collector-opentelemetry-collector.observability:4317", insecure=True)
    trace.get_tracer_provider().add_span_processor(
        BatchSpanProcessor(otel_exporter)
    )

    RequestsInstrumentor().instrument(tracer_provider=trace.get_tracer_provider())
    tracer = trace.get_tracer(__name__)
    return tracer


