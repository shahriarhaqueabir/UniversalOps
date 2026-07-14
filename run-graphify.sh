#!/bin/sh
curl -s http://localhost:11434/api/generate \
  -d '{"model":"graphify:latest","prompt":"Say hello","stream":false}'