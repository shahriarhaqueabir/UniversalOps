import urllib.request
import json

data = json.dumps({
    "model": "graphify:latest",
    "prompt": "Say hello and confirm you are running",
    "stream": False
}).encode("utf-8")

req = urllib.request.Request(
    "http://localhost:11434/api/generate",
    data=data,
    headers={"Content-Type": "application/json"}
)

try:
    resp = urllib.request.urlopen(req, timeout=120)
    result = json.loads(resp.read().decode("utf-8"))
    print(result.get("response", "NO RESPONSE"))
except Exception as e:
    print(f"ERROR: {e}")
