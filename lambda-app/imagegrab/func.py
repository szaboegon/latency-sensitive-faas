from parliament import Context
from flask import Request
import json
import cv2
import numpy as np
import os
import requests
import base64
from concurrent.futures import ThreadPoolExecutor
from opentelemetry.propagate import inject, extract
import tracing
    
def main(context: Context):
    tracer = tracing.instrument_app()
    with tracer.start_as_current_span("start_imagegrab", context=extract(context.request.headers)) as span:
        return handler(context=context)
   
def handler(context: Context):
    image_bytes = context.request.data
    base64_image = image_bytes.decode('utf-8')
    event_out = {
        "image": base64_image,
        "original_image": base64_image
    }
    if event_out is not None:
        # TODO replace this later with an event omit possibly
        headers = {}
        inject(headers)
        resp = requests.post("http://resize.default.svc.cluster.local", json=event_out, headers=headers)
        return resp.text, 200
    
    return "{'message': 'No image found'}", 400

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")
 
def fire_and_forget(url, json=None):
    with ThreadPoolExecutor() as executor:
        future = executor.submit(requests.post, url, json=json)


    