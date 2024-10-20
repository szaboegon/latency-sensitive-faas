from parliament import Context
from flask import Request
import json
import os
from imagegrab import handler as imagegrab
from resize import handler as resize

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    match forward_to:
        case "imagegrab": 
            return imagegrab(context)
        case "resize":
            return resize(context)


