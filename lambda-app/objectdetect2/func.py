from parliament import Context
from flask import Request
import json
import tracing
import numpy as np
import cv2
import base64
import requests
from opentelemetry.propagate import inject, extract

#TODO to env variable
CONFIDENCE_MIN = 0.4

def main(context: Context):
    tracer = tracing.instrument_app()
    with tracer.start_as_current_span("start_objectdetect2", context=extract(context.request.headers)) as span:
        return handler(context=context)
    
def handler(context: Context):
    json_data = context.request.json
     # Convert image from string
    image = base64_to_image(json_data.get("image"))

    (h, w) = image.shape[:2]
    blob = cv2.dnn.blobFromImage(image, 0.007843, (h, w), 127.5)

    net.setInput(blob)
    detections = net.forward()
    best_confidence = 0
    label = "no object"

    for i in np.arange(0, detections.shape[2]):
        confidence = detections[0, 0, i, 2]
        if confidence > CONFIDENCE_MIN:
            idx = int(detections[0, 0, i, 1])
            # display the prediction
            if best_confidence < confidence:
                best_confidence = confidence
                label = "{}: {:.2f}%".format(CLASSES[idx], confidence * 100)

    label = f"IMG-{json_data.get('cropped_img_id')}: "\
        f"STAGE-1: {json_data.get('cropped_coords')['label']['name']}; "\
        f"STAGE-2: {label}"

    event_out = {"label": label,
                 "cropped_img_id": json_data.get('cropped_img_id'),
                 "cropped_coords": json_data.get("cropped_coords"),
                 "original_image": json_data.get("original_image")}
    
    headers = {}
    inject(headers)
    resp = requests.post("http://tag.default.svc.cluster.local", json=event_out, headers=headers)
    return resp.text, 200

CLASSES = ["background", "aeroplane", "bicycle", "bird", "boat",
           "bottle", "bus", "car", "cat", "chair", "cow", "diningtable",
           "dog", "horse", "motorbike", "person", "pottedplant", "sheep",
           "sofa", "train", "tvmonitor"]

# Load serialized model from disk
net = cv2.dnn.readNetFromCaffe("./MobileNetSSD_deploy.prototxt.txt",
                               "./MobileNetSSD_deploy.caffemodel")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)