from parliament import Context
import base64
import cv2
import numpy as np

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    decoded = cv2.imdecode(image, flags=1)
    if decoded is None:
        raise ValueError("Decoded image is None. Base64 data may be invalid.")
    return decoded

def handler(event):
    json_data = event.json
     # Convert image from base64
    image = base64_to_image(json_data.get("image"))

    # Grayscale image
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)

    # Trigger object detection function
    event_out = {"image": image_to_base64(image),
                 "original_image": json_data.get("original_image"),
                 "origin_h": json_data.get("origin_h"),
                 "origin_w": json_data.get("origin_w")}
    
    return event_out
