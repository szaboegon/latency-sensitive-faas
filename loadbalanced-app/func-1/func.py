from parliament import Context
import os
import requests
from imagegrab import handler as imagegrab
from resize import handler as resize

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
def headers(component):
    return {
    "X-Forward-To": component,
    "Content-Type": "application/json"
    }

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    next_component = ""
    event_out = {}
    match forward_to:
        case "imagegrab": 
            next_component, event_out = imagegrab(context)
        case "resize":
            next_component, event_out = resize(context)

    if next_component != "":
        resp = requests.post(LOADBALANCER_URL, json=event_out, headers=headers(next_component))
        return resp.text, 200
    else:
        return "ok", 200


