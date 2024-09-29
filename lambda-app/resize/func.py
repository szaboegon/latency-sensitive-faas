from parliament import Context
from flask import Request
import json
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
    # Convert image from base64
    json_data = context.request.json
    image = base64_to_image(json_data.get("image"))

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
                "original_image": json_data.get("original_image"),
                "origin_h": h,
                "origin_w": w}
    
    resp = requests.post("http://grayscale.default.svc.cluster.local", json=event_out)

    return resp.text, 200