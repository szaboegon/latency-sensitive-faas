from parliament import Context
from flask import Request
import base64
import cv2
import numpy as np
import requests

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def main(context: Context):
     # Convert image from base64
    image = base64_to_image(context.request.get("image"))

    # Grayscale image
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)

    # Trigger object detection function
    event_out = {"image": image_to_base64(image),
                 "origin_location": context.request["origin_location"],
                 "origin_h": context.request["origin_h"],
                 "origin_w": context.request["origin_w"]}
    
    requests.post("http://objectdetect.default.svc.cluster.local", json=event_out)
