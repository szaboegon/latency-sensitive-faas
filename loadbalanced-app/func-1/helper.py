import os

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080'
def headers(component):
    return {
    "X-Forward-To": component,
    "Content-Type": "application/json"
    }
