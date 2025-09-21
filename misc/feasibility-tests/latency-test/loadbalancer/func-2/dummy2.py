import time

def handler(context):
    time.sleep(1)
    return "dummy3", context.request.data