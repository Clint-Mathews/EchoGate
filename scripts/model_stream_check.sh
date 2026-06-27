#!/bin/sh

echo "Checking if model is running..."

# Docummentation for reference: https://github.com/ollama/ollama/blob/main/docs/api.md#generate-request-streaming
curl -s http://localhost:11434/api/generate -d '{
  "model": "llama3.2",
  "prompt": "Why is the sky blue?"
}' | jq .
