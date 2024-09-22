from parliament import Context
from flask import Request
import json
import cv2
import numpy as np
import os
import requests
import base64

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def process_image(image):
    # cap = cv2.VideoCapture(picture)
    # ret, image = cap.read()
    
    if not np.any(np.equal(image, None)):
        # Trigger resize function
        event_out = {"image": image_to_base64(image),
                     "original_image": image_to_base64(image)} #TODO the original image (without modifications), this should be renamed and readjusted in all following functions
        return event_out   
    else:
        return None
    
def main(context: Context):
    base64 = context.request.data
    image = base64_to_image(base64)
    event_out = process_image(image)

    if event_out is not None:
        # TODO replace this later with an event omit possibly
        requests.post("http://resize.default.svc.cluster.local", json=event_out)
        return "{}", 200
    
    return "{'message': 'No image found'}", 400
    

    