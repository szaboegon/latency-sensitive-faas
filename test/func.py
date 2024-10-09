from parliament import Context
from flask import Request
import requests
import json
import time
import uuid
import os

def main(context: Context):
    #node_ip = os.environ['NODE_IP']
    # OTLP HTTP endpoint (replace with your collector's OTLP HTTP endpoint)
    otlp_endpoint = f"http://otel-collector-opentelemetry-collector.observability:4318/v1/traces"
    apm_endpoint = "http://apm-server-apm-server.observability.svc.cluster.local:8200"

    # Trace and Span ID generation (must be 16-byte hexadecimal for span_id, 32-byte for trace_id)
    trace_id = uuid.uuid4().hex  # 32-byte hex
    span_id = uuid.uuid4().hex[:16] 

    # Current timestamp in nanoseconds
    current_time_ns = int(time.time() * 1e9)

    # Construct the OTLP trace payload in JSON format
    trace_payload = {
        "resource_spans": [
            {
                "resource": {
                    "attributes": [
                        {
                            "key": "service.name",
                            "value": {"stringValue": "manual-otlp-trace"}
                        }
                    ]
                },
                "instrumentationLibrarySpans": [
                    {
                       "instrumentationLibrary": {
                        "name": "manual-trace-lib",
                        "version": "1.0"
                    },
                        "spans": [
                            {
                                "traceId": trace_id,
                                "spanId": span_id,
                                "name": "hello-world-span",
                                "kind": 1,  # Internal span kind
                                "startTimeUnixNano": current_time_ns,
                                "endTimeUnixNano": current_time_ns + 1000000,  # 1ms duration
                                "attributes": [
                                    {
                                        "key": "message",
                                        "value": {"stringValue": "Hello, World!"}
                                    }
                                ]
                            }
                        ]
                    }
                ]
            }
        ]
    }

    # Send the HTTP POST request with the trace data
    headers = {'Content-Type': 'application/json'}
    response = requests.post(apm_endpoint, headers=headers, data=json.dumps(trace_payload))

    # Check the response from the OpenTelemetry collector
    if response.status_code == 200:
        return "Trace sent successfully!", 200
    else:
        return f"Failed to send trace. Status code: {response.status_code}, Response: {response.text}", 400
