#!/usr/bin/env python3

### IMAGEGRAB application component

import cv2
import numpy as np
import os

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def process_image(picture, output_function, context):
    cap = cv2.VideoCapture(picture)
    ret, image = cap.read()
    if not np.any(np.equal(image, None)):
        # Trigger resize function
        event_out = {"image": image_to_base64(image),
                     "origin_location": picture}
        output_function.call(event_out, context)
    else:
        print("No image found")
        return False

def IMAGEGRAB_handler(event, context):
    picture = event.get("picture")
    process_image(picture,
                  context.downstream_functions.RESIZE,
                  context)

### RESIZE application component

import base64
import cv2
import numpy as np

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def RESIZE_handler(event, context):
    # Convert image from base64
    image = base64_to_image(event.get("image"))

    # Resize image
    (h, w) = image.shape[:2]

    # Resize image
    scale_percent = int(context.env_vars.SCALE_PERCENT)  # percent of original size
    width = int(image.shape[1] * scale_percent / 100)
    height = int(image.shape[0] * scale_percent / 100)
    dim = (width, height)

    image = cv2.resize(image, dim, interpolation=cv2.INTER_AREA)

    # Trigger grayscale function
    event_out = {"image": image_to_base64(image),
                 "origin_location": event["origin_location"],
                 "origin_h": h,
                 "origin_w": w}
    context.downstream_functions.GRAYSCALE.call(event_out, context)

### GRAYSCALE application component

import base64
import cv2
import numpy as np

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def GRAYSCALE_handler(event, context):
    # Convert image from base64
    image = base64_to_image(event.get("image"))

    # Grayscale image
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)

    # Trigger object detection function
    event_out = {"image": image_to_base64(image),
                 "origin_location": event["origin_location"],
                 "origin_h": event["origin_h"],
                 "origin_w": event["origin_w"]}
    context.downstream_functions.OBJECTDETECT.call(event_out, context)

### OBJECTDETECT application component

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

def OBJECTDETECT_handler(event, context):
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

### CUT application component

import numpy as np

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)

def CUT_handler(event, context):
    if 'origin_location' in event and 'cropped_coords' in event:
        # Log object count
        print("Object count:", len(event['cropped_coords']))
        # Read original image and perform the crop on that
        cap = cv2.VideoCapture(event["origin_location"])
        ret, image = cap.read()
        img_id = 1
        for i in event['cropped_coords']:
            # call object detection for the cropped image
            cropped_image = image[i["startY"]:i["endY"],
                                  i["startX"]:i["endX"]]
            event_out = {"image":
                         image_to_base64(cropped_image),
                         "cropped_img_id": img_id,
                         "cropped_coords": i,
                         "origin_location": event["origin_location"]}
            img_id += 1
            context.downstream_functions.OBJECTDETECT2.call(event_out, context)
    else:
        print("No regions to cut")

### OBJECTDETECT2 application component

import numpy as np
import cv2

CLASSES = ["background", "aeroplane", "bicycle", "bird", "boat",
           "bottle", "bus", "car", "cat", "chair", "cow", "diningtable",
           "dog", "horse", "motorbike", "person", "pottedplant", "sheep",
           "sofa", "train", "tvmonitor"]

# Load serialized model from disk
net = cv2.dnn.readNetFromCaffe("./MobileNetSSD_deploy.prototxt.txt",
                               "./MobileNetSSD_deploy.caffemodel")


def OBJECTDETECT2_handler(event, context):
    # Convert image from string
    image = base64_to_image(event.get("image"))

    (h, w) = image.shape[:2]
    blob = cv2.dnn.blobFromImage(image, 0.007843, (h, w), 127.5)

    net.setInput(blob)
    detections = net.forward()
    best_confidence = 0
    label = "no object"

    for i in np.arange(0, detections.shape[2]):
        confidence = detections[0, 0, i, 2]
        if confidence > context.env_vars.CONFIDENCE_MIN:
            idx = int(detections[0, 0, i, 1])
            # display the prediction
            if best_confidence < confidence:
                best_confidence = confidence
                label = "{}: {:.2f}%".format(CLASSES[idx], confidence * 100)

    label = f"IMG-{event['cropped_img_id']}: "\
        f"STAGE-1: {event['cropped_coords']['label']['name']}; "\
        f"STAGE-2: {label}"

    event_out = {"label": label,
                 "cropped_img_id": event['cropped_img_id'],
                 "cropped_coords": event["cropped_coords"],
                 "origin_location": event["origin_location"]}
    context.downstream_functions.TAG.call(event_out, context)

### TAG application component

import cv2
import numpy as np

# Generate a set of bounding box colors for each class
CLASSES = ["background", "aeroplane", "bicycle", "bird", "boat",
           "bottle", "bus", "car", "cat", "chair", "cow", "diningtable",
           "dog", "horse", "motorbike", "person", "pottedplant", "sheep",
           "sofa", "train", "tvmonitor"]
COLORS = np.random.uniform(0, 255, size=(len(CLASSES), 3))

def TAG_handler(event, context):
    # Read original image and perform the crop on that
    cap = cv2.VideoCapture(event["origin_location"])
    ret, image = cap.read()

    # Tag image
    label = event["label"]
    (h, w) = image.shape[:2]
    index = event["cropped_coords"]["label"]["index"]
    startY = event["cropped_coords"]["startY"]
    cv2.rectangle(image,
                  (event["cropped_coords"]["startX"],
                   startY),
                  (event["cropped_coords"]["endX"],
                   event["cropped_coords"]["endY"]),
	    	  COLORS[index], 2)
    y = startY - 15 if startY - 15 > 15 else startY + 15
    cv2.putText(image, label, (event["cropped_coords"]["startX"], y),
	    	cv2.FONT_HERSHEY_SIMPLEX, 0.5,
                COLORS[index], 2)
    out_file = f"result-{event['cropped_img_id']}.jpg"
    print(out_file)
    cv2.imwrite(out_file, image)

### A simplification of the Wrapper for function calls and env vars

class EnvVars:
    def __init__(self):
        pass

class DownstreamFunctions:
    def __init__(self):
        pass

class FunctionCall:
    def __init__(self, function_name):
        self.function_name = function_name

    def call(self, event, context):
        globals()[self.function_name](event, context)

class Context:
    env_vars = EnvVars()
    downstream_functions = DownstreamFunctions()

    def __init__(self):
        setattr(self.env_vars, "SCALE_PERCENT", 25)
        setattr(self.env_vars, "CONFIDENCE_MIN", 0.4)
        setattr(self.downstream_functions, "RESIZE",
                FunctionCall("RESIZE_handler"))
        setattr(self.downstream_functions, "GRAYSCALE",
                FunctionCall("GRAYSCALE_handler"))
        setattr(self.downstream_functions, "OBJECTDETECT",
                FunctionCall("OBJECTDETECT_handler"))
        setattr(self.downstream_functions, "CUT",
                FunctionCall("CUT_handler"))
        setattr(self.downstream_functions, "OBJECTDETECT2",
                FunctionCall("OBJECTDETECT2_handler"))
        setattr(self.downstream_functions, "TAG",
                FunctionCall("TAG_handler"))

### Entry point

if __name__ == '__main__':
    context = Context()

    IMAGEGRAB_handler({"picture": "./test-1.jpg"}, context)
