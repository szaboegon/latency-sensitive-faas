from parliament import Context
from flask import Request
import json
import cv2
import numpy as np
import base64
    
def handler(context: Context):
    json_data = context.request.json
    # Read original image and perform the crop on that
    image = base64_to_image(json_data.get("original_image"))

    # Tag image
    label = json_data["label"]
    (h, w) = image.shape[:2]
    index = json_data["cropped_coords"]["label"]["index"]
    startY = json_data["cropped_coords"]["startY"]
    cv2.rectangle(image,
                  (json_data["cropped_coords"]["startX"],
                   startY),
                  (json_data["cropped_coords"]["endX"],
                   json_data["cropped_coords"]["endY"]),
	    	  COLORS[index], 2)
    y = startY - 15 if startY - 15 > 15 else startY + 15
    cv2.putText(image, label, (json_data["cropped_coords"]["startX"], y),
	    	cv2.FONT_HERSHEY_SIMPLEX, 0.5,
                COLORS[index], 2)
    out_file = f"result-{json_data['cropped_img_id']}.jpg"
    print(out_file)
    cv2.imwrite(out_file, image)
    return "Image tag was successful", 200

# Generate a set of bounding box colors for each class
CLASSES = ["background", "aeroplane", "bicycle", "bird", "boat",
           "bottle", "bus", "car", "cat", "chair", "cow", "diningtable",
           "dog", "horse", "motorbike", "person", "pottedplant", "sheep",
           "sofa", "train", "tvmonitor"]
COLORS = np.random.uniform(0, 255, size=(len(CLASSES), 3))

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)