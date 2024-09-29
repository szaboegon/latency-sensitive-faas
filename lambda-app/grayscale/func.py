from parliament import Context
from flask import Request
import base64
import cv2
import numpy as np
import requests
import tracing

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def main(context: Context):
        tracer = tracing.instrument_app()
        with tracer.start_as_current_span("start") as span:
            return handler(context=context)

def handler(context: Context):
    json_data = context.request.json
     # Convert image from base64
    image = base64_to_image(json_data.get("image"))

    # Grayscale image
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)

    # Trigger object detection function
    event_out = {"image": image_to_base64(image),
                 "original_image": json_data.get("original_image"),
                 "origin_h": json_data.get("origin_h"),
                 "origin_w": json_data.get("origin_w")}
    
    resp = requests.post("http://objectdetect.default.svc.cluster.local", json=event_out)
    return resp.text, 200
