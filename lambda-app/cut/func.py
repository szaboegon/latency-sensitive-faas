from parliament import Context
from flask import Request
import json
import tracing
import numpy as np
import cv2
import requests
import base64

def main(context: Context):
    tracer = tracing.instrument_app()
    with tracer.start_as_current_span("start") as span:
        return handler(context=context)

def handler(context: Context):
    json_data = context.request.json
    if 'original_image' in json_data and 'cropped_coords' in json_data:
        # Log object count
        print("Object count:", len(json_data.get('cropped_coords')))
        # Read original image and perform the crop on that
        image = base64_to_image(json_data.get("original_image"))
        img_id = 1
        for i in json_data.get('cropped_coords'):
            # call object detection for the cropped image
            cropped_image = image[i["startY"]:i["endY"],
                                  i["startX"]:i["endX"]]
            event_out = {"image":
                         image_to_base64(cropped_image),
                         "cropped_img_id": img_id,
                         "cropped_coords": i,
                         "original_image": json_data.get("original_image")}
            img_id += 1

            resp = requests.post("http://objectdetect2.default.svc.cluster.local", json=event_out)
            return resp.text, 200
    else:
        return "Invalid inputs", 400

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)