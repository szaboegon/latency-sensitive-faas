from parliament import Context
from flask import Request
import json
import base64
import cv2
import numpy as np

# Initialize the list of class labels MobileNet SSD was trained to detect
CLASSES = ["background", "aeroplane", "bicycle", "bird", "boat",
           "bottle", "bus", "car", "cat", "chair", "cow", "diningtable",
           "dog", "horse", "motorbike", "person", "pottedplant", "sheep",
           "sofa", "train", "tvmonitor"]

# Load serialized model from disk
print("[INFO] loading model...")
net = cv2.dnn.readNetFromCaffe("./MobileNetSSD_deploy.prototxt.txt",
                               "./MobileNetSSD_deploy.caffemodel")

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def detect_objects(image, origin_h, origin_w, confidence_min):
    (h, w) = image.shape[:2]
    blob = cv2.dnn.blobFromImage(image, 0.007843, (h, w), 127.5)

    net.setInput(blob)
    detections = net.forward()

    labels = []
    for i in np.arange(0, detections.shape[2]):
        confidence = detections[0, 0, i, 2]
        if confidence > confidence_min:
            idx = int(detections[0, 0, i, 1])

            # Mark area on the original-sized picture not the resized
            box = detections[0, 0, i, 3:7] * np.array([origin_w,
                                                       origin_h,
                                                       origin_w,
                                                       origin_h])
            (startX, startY, endX, endY) = box.astype("int")

            labels.insert(0, {"startX": int(startX),
                              "startY": int(startY),
                              "endX": int(endX),
                              "endY": int(endY),
                              "label": {"name": CLASSES[idx],
                                        "index": int(idx)},
                              "confidence": float(confidence)})
    return labels

def main(context: Context):
    #TODO
    event = context.request
     # Convert image from string
    image = base64_to_image(event.get("image"))

    origin_h, origin_w = int(event["origin_h"]), int(event["origin_w"])

    # Detect objects
    coords = detect_objects(image, origin_h, origin_w,
                            context.env_vars.CONFIDENCE_MIN)

    # Trigger cut function
    event_out = {"cropped_coords": coords,
                 "origin_location": event["origin_location"],}
    context.downstream_functions.CUT.call(event_out, context)

