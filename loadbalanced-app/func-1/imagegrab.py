from parliament import Context
from flask import Request
import json
import cv2
import numpy as np
import os
import requests
import base64
from concurrent.futures import ThreadPoolExecutor
from helper import LOADBALANCER_URL

def handler(context: Context):
    image_bytes = context.request.data
    base64_image = image_bytes.decode('utf-8')
    event_out = {
        "image": base64_image,
        "original_image": base64_image
    }
    if event_out is not None:
        # TODO replace this later with an event omit possibly
        resp = requests.post(f"{LOADBALANCER_URL}/resize", json=event_out)
        return resp.text, 200
    
    return "{'message': 'No image found'}", 400

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")


    