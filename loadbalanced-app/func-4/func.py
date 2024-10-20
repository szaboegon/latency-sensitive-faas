from parliament import Context
from flask import Request
import json
from objectdetect import handler as objectdetect

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    match forward_to:
        case "objectdetect": 
            return objectdetect(context)