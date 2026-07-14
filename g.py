import json
import sys
import urllib.request

d = json.dumps(
    {"model": "graphify:latest", "prompt": "hello", "stream": False}
).encode()
r = urllib.request.urlopen(
    urllib.request.Request(
        "http://localhost:11434/api/generate", d, {"Content-Type": "application/json"}
    ),
    120,
)
print(json.loads(r.read())["response"][:200])
