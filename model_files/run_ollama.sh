#!/bin/sh

# 1. Start Ollama in the background
echo "Starting Ollama server..."
ollama serve &

# Save the process ID of the Ollama server so we can wait on it later
pid=$!

# 2. Wait for Ollama to start accepting connections
echo "Waiting for Ollama server to be fully ready..."
until ollama list > /dev/null 2>&1; do
  sleep 1
done

echo "Ollama server is up!"

# 3. Pull the model (use 'pull' instead of 'run' in headless environments)
echo "Downloading llama3.2 model..."
ollama pull llama3.2

echo "Model downloaded successfully! Keeping container alive..."

# 4. Wait on the background Ollama process so the container stays running
wait $pid
