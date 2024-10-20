from parliament import Context
from flask import Request
import base64
import cv2
import numpy as np
import requests
from helper import LOADBALANCER_URL

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

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
    
    resp = requests.post(f"{LOADBALANCER_URL}/objectdetect", json=event_out)
    return resp.text, 200
