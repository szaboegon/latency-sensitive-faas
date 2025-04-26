from parliament import Context
import cv2
import base64

def handler(event):
    image_bytes = event.data
    base64_image = image_bytes.decode('utf-8')
    event_out = {
        "image": base64_image,
        "original_image": base64_image
    }
    if event_out is not None:
        return event_out
    
    raise Exception("no image was found")

def image_to_base64(image):
    retval, buffer = cv2.imencode('.jpg', image)
    return base64.b64encode(buffer).decode("utf-8")


    