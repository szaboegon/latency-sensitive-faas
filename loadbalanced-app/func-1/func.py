from parliament import Context
import os
import requests
from imagegrab import handler as imagegrab
from resize import handler as resize
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
if 'tracer' not in globals():
    tracer = tracing.instrument_app("func-1")

def get_headers(component):
    return {
    "X-Forward-To": component,
    "Content-Type": "application/json"
    }

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    link = trace.Link(trace.get_current_span(extract(context.request.headers)).get_span_context())
    with tracer.start_as_current_span(forward_to, links=[link]) as span:
        next_component = ""
        event_out = {}
        match forward_to:
            case "imagegrab": 
                next_component, event_out = imagegrab(context)
            case "resize":
                next_component, event_out = resize(context)

        if next_component != "":
            headers = get_headers(next_component)
            inject(headers)
            resp = requests.post(LOADBALANCER_URL, json=event_out, headers=headers)
            return resp.text, 200
        else:
            return "ok", 200


