from parliament import Context
from flask import Request
import os

def main(context: Context):
    value = os.environ["key-1"]
    return value, 200
