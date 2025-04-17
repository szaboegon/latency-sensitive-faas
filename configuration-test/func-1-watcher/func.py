from parliament import Context
from flask import Request
import json
import os
import threading
import time
import logging

def load_configmap(path: str) -> dict:
    data = {}

    for root, dirs, files in os.walk(path):
        for name in files:
            file_path = os.path.join(root, name)

            try:
                # Resolve symlinks
                while os.path.islink(file_path):
                    link_target = os.readlink(file_path)
                    if not os.path.isabs(link_target):
                        file_path = os.path.join(os.path.dirname(file_path), link_target)
                    else:
                        file_path = link_target

                if os.path.isdir(file_path):
                    continue

                with open(file_path, "r") as f:
                    data[name] = f.read()
            except Exception as e:
                raise RuntimeError(f"Failed reading {file_path}: {e}")

    return data

config = None
def refresh():
    global config
    while True:
        config_dir = "/workspace/configmap"
        config = load_configmap(config_dir)
        logging.warning("refreshed")
        time.sleep(1)

def main(context: Context):
    config_dir = "/workspace/configmap"
    config = load_configmap(config_dir)
    t = threading.Thread(target=refresh, daemon=True)
    t.start()
    return config["key-1"], 200

