from parliament import Context
from flask import Request
from objectdetect import handler as objectdetect
from objectdetect2 import handler as objectdetect2

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    match forward_to:
        case "objectdetect": 
            return objectdetect(context)
        case "objectdetect2":
            return objectdetect2(context)