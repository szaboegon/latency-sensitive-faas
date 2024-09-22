from parliament import Context
from flask import Request
import json
import cv2
import numpy as np
import os
import requests
import base64

def main(context: Context):
    picture = context.request.get("picture")
    process_image(picture)


def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def process_image(picture):
    cap = cv2.VideoCapture(picture)
    ret, image = cap.read()
    if not np.any(np.equal(image, None)):
        # Trigger resize function
        event_out = {"image": image_to_base64(image),
                     "origin_location": picture}
        # TODO replace this later with an event omit possibly
        requests.post("http://resize.default.svc.cluster.local", json=event_out)
    else:
        print("No image found")
        return False
    