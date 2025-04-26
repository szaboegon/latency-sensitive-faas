from parliament import Context
import numpy as np
import cv2
import base64

def handler(event):
    json_data = event.json
    if 'original_image' in json_data and 'cropped_coords' in json_data:
        # Log object count
        print("Object count:", len(json_data.get('cropped_coords')))
        # Read original image and perform the crop on that
        image = base64_to_image(json_data.get("original_image"))
        img_id = 1
        for i in json_data.get('cropped_coords'):
            # call object detection for the cropped image
            cropped_image = image[i["startY"]:i["endY"],
                                  i["startX"]:i["endX"]]
            event_out = {"image":
                         image_to_base64(cropped_image),
                         "cropped_img_id": img_id,
                         "cropped_coords": i,
                         "original_image": json_data.get("original_image")}
            img_id += 1

            return event_out
    else:
        return "Invalid inputs", 400

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")

def base64_to_image(text):
    image = base64.b64decode(text)
    image = np.frombuffer(image, dtype=np.uint8)
    return cv2.imdecode(image, flags=1)