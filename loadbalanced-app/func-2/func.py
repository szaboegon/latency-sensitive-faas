from parliament import Context
from flask import Request
from cut import handler as cut
from grayscale import handler as grayscale
import os

def main(context: Context):
    forward_to = context.request.headers.get("X-Forward-To")
    match forward_to:
        case "grayscale": 
            return grayscale(context)
        case "cut":
            return cut(context)