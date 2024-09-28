from parliament import Context
from flask import Request
import base64
import cv2
import numpy as np
import requests
from opentelemetry import trace
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.baggage.propagation import W3CBaggagePropagator
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
from opentelemetry.instrumentation.requests import RequestsInstrumentor

trace.set_tracer_provider(
TracerProvider(
    resource=Resource.create({SERVICE_NAME: "grayscale"})
)
)
jaeger_exporter = JaegerExporter(
collector_endpoint="http://jaeger-collector.observability.svc.cluster.local:14268/api/traces"
)

trace.get_tracer_provider().add_span_processor(
BatchSpanProcessor(jaeger_exporter)
)
tracer = trace.get_tracer(__name__)
RequestsInstrumentor().instrument(tracer_provider=trace.get_tracer_provider())

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def main(context: Context):
     # Convert image from base64
    return {}, 200
    image = base64_to_image(context.request.get("image"))

    # Grayscale image
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)

    # Trigger object detection function
    event_out = {"image": image_to_base64(image),
                 "origin_location": context.request["origin_location"],
                 "origin_h": context.request["origin_h"],
                 "origin_w": context.request["origin_w"]}
    
    requests.post("http://objectdetect.default.svc.cluster.local", json=event_out)
