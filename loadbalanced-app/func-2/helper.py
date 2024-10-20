import os

LOADBALANCER_URL = f'http://{os.environ["NODE_IP"]}:8080/forward'