from parliament import Context
import os
import requests
from objectdetect import handler as objectdetect
from opentelemetry.propagate import inject, extract
import tracing
from opentelemetry import trace

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
if 'tracer' not in globals():
    tracer = tracing.instrument_app("func-4")
    
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

def handle_request(context, parent_context):
    """
    Handles the request by invoking the appropriate handler and preparing
    the next component's details.
    """
    forward_to = context.request.headers.get("X-Forward-To")
    if not forward_to:
        return "", {}

    with tracer.start_as_current_span(forward_to, context=parent_context) as span:
        match forward_to:
            case "objectdetect":
                return objectdetect(context)
            case _:
                return "", {}

def forward_request(next_component, event_out, parent_context):
    """
    Forwards the request to the load balancer if there is a next component.
    """
    if not next_component:
        return "ok", 200

    headers = get_headers(next_component, parent_context)
    response = requests.post(LOADBALANCER_URL, json=event_out, headers=headers)
    return response.text, 200

def main(context: Context):
    """
    Entry point of the function. Manages parent context and orchestrates the pipeline.
    """
    parent_context = extract(context.request.headers)
    if not parent_context:
        with tracer.start_as_current_span("objectdetect_pipeline") as parent_span:
            next_component, event_out = handle_request(context, trace.set_span_in_context(parent_span))
            return forward_request(next_component, event_out, trace.set_span_in_context(parent_span))
    else:
        next_component, event_out = handle_request(context, parent_context)
        return forward_request(next_component, event_out, parent_context)

    
