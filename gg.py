import json
import sys
import urllib.request

prompt_file = sys.argv[1] if len(sys.argv) > 1 else None

if prompt_file:
    with open(prompt_file, "r", encoding="utf-8") as f:
        prompt = f.read()
else:
    prompt = "Say hello"

data = json.dumps(
    {"model": "graphify:latest", "prompt": prompt, "stream": False}
).encode("utf-8")

req = urllib.request.Request(
    "http://localhost:11434/api/generate",
    data=data,
    headers={"Content-Type": "application/json"},
)

try:
    resp = urllib.request.urlopen(req, timeout=180)
    result = json.loads(resp.read().decode("utf-8"))
    print(result.get("response", "NO RESPONSE"))
except Exception as e:
    print("ERROR: " + str(e))
