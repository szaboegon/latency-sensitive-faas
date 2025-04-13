from parliament import Context
import os
import requests
from dummy1 import handler as dummy1
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace
import threading

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
HANDLERS = {
    "dummy1": dummy1,
}

if 'tracer' not in globals():
    tracer = tracing.instrument_app("func-1-direct")
    
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
    headers = get_headers(next_component, span_context=span_context)
    #threading.Thread(target=send, args=(event_out, headers))
    send(event_out, headers)
    return "ok", 200

def send(event_out, headers):
    requests.post("http://func-2-direct.application.svc.cluster.local", data=event_out, headers=headers)

def main(context: Context):
    """
    Entry point of the function. Manages parent context and orchestrates the pipeline.
    """
    next_component, event_out, span_context = handle_request(context)
    return forward_request(next_component, event_out, span_context)

    
