from parliament import Context
from flask import Request
import json
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
    resource=Resource.create({SERVICE_NAME: "resize"})
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
    data = context.request.data
    image = base64_to_image(data.get("image"))

    # Resize image
    (h, w) = image.shape[:2]

    # Resize image
    scale_percent = int(25)  # percent of original size
    width = int(image.shape[1] * scale_percent / 100)
    height = int(image.shape[0] * scale_percent / 100)
    dim = (width, height)

    image = cv2.resize(image, dim, interpolation=cv2.INTER_AREA)

    # Trigger grayscale function
    event_out = {"image": image_to_base64(image),
                "original_image": data.get("original_image"),
                "origin_h": h,
                "origin_w": w}
    
    resp = requests.post("http://grayscale.default.svc.cluster.local", data=event_out)

    return '{"message":"ok from resize"}', 200