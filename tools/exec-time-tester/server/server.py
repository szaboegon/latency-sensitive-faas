from flask import Flask, request, jsonify
import importlib

# Dynamically import your function
module = importlib.import_module("main")  # assumes your function lives in main.py
handler = getattr(module, "handler")

app = Flask(__name__)

@app.route("/function", methods=["POST"])
def handle():
    event = request  # you can wrap this in a custom Context if needed
    result = handler(event)
    return jsonify(result)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
