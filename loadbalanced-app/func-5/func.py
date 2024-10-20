from parliament import Context
from flask import Request
import json
from objectdetect2 import handler as objectdetect2
from tag import handler as tag

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    match forward_to:
        case "objectdetect2": 
            return objectdetect2(context)
        case "tag": 
            return tag(context)