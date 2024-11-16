from parliament import Context
import os
import requests
from tag import handler as tag
from objectdetect2 import handler as objectdetect2
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
HANDLERS = {
    "objectdetect2": objectdetect2,
    "tag": tag,
}

if 'tracer' not in globals():
    tracer = tracing.instrument_app("func-5")
    
def get_headers(component, span_context=None):
    """
    Generates headers for the next component, including trace context.
    """
    headers = {
        "X-Forward-To": component,
        "Content-Type": "application/json"
    }
    if span_context:
        inject(headers, context=span_context)
    return headers

def handle_request(context):
    """
    Handles the request by invoking the appropriate handler and preparing
    the next component's details.
    """
    forward_to = context.request.headers.get("X-Forward-To")
    if not forward_to:
        return "", {}, None

    with tracer.start_as_current_span(forward_to, context=extract(context.request.headers)) as span:
        if forward_to in HANDLERS:
            next_component, event_out = HANDLERS[forward_to](context)
            return next_component, event_out, trace.set_span_in_context(span)
        return None, {}, trace.set_span_in_context(span)

def forward_request(next_component, event_out, span_context):
    """
    Forwards the request to the load balancer if there is a next component.
    """
    if not next_component:
        return "ok", 200

    headers = get_headers(next_component, span_context)
    response = requests.post(LOADBALANCER_URL, json=event_out, headers=headers)
    return response.text, 200

def main(context: Context):
    """
    Entry point of the function. Manages parent context and orchestrates the pipeline.
    """
    next_component, event_out, span_context = handle_request(context)
    return forward_request(next_component, event_out, span_context)

    
