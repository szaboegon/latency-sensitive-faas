from parliament import Context
import os
import requests
from objectdetect2 import handler as objectdetect2
from tag import handler as tag
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
def get_headers(component):
    return {
    "X-Forward-To": component,
    "Content-Type": "application/json"
    }

if 'tracer' not in globals():
    tracer = tracing.instrument_app("func-5")

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    with tracer.start_as_current_span(forward_to, links=trace.Link(extract(context.request.headers))) as span:
        next_component = ""
        event_out = {}
        match forward_to:
            case "objectdetect2": 
                next_component, event_out = objectdetect2(context)
            case "tag": 
                next_component, event_out = tag(context)

        if next_component != "":
            headers = get_headers(next_component)
            inject(headers)
            resp = requests.post(LOADBALANCER_URL, json=event_out, headers=headers)
            return resp.text, 200
        else:
            return "ok", 200

