import time

def handler(context):
    time.sleep(1)
    return "dummy2", context.request.data