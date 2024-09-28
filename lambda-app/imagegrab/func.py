from parliament import Context
from flask import Request
import json
import cv2
import numpy as np
import os
import requests
import base64
from concurrent.futures import ThreadPoolExecutor
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
    resource=Resource.create({SERVICE_NAME: "imagegrab"})
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
 
def fire_and_forget(url, json=None):
    with ThreadPoolExecutor() as executor:
        future = executor.submit(requests.post, url, json=json)
    
def main(context: Context):
    data = context.request.data
    event_out = {
        "image": data,
        "original_image": data
    }

    if event_out is not None:
        # TODO replace this later with an event omit possibly
        #fire_and_forget("http://resize.default.svc.cluster.local", json=event_out)
        #resp = requests.post("http://imagegrab.default.127.0.0.1.sslip.io", json=event_out)
        resp = requests.post("http://resize.default.svc.cluster.local", data=event_out)
        return resp.text, 200
    
    return "{'message': 'No image found'}", 400


    