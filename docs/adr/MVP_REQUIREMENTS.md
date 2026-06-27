# EchoGate: MVP Requirements Specification

EchoGate is a high-performance, split-plane AI API Gateway and Reverse Proxy engineered in Golang and Python to optimize upstream LLM traffic. It intercepts prompt streams to deliver lightning-fast semantic caching and token accounting, dropping p99 latencies and inference costs before visualizing cluster health through a real-time React dashboard.

---

## 1. System Topology Overview

The MVP architecture operates across four distinct layers running locally:
* **Local LLM Engine:** Ollama serving models via an OpenAI-compatible API.
* **The Data Plane (Golang):** High-throughput, low-latency HTTP reverse proxy handling real-time token streaming.
* **The Control Plane (Python/Flask):** Intelligence layer managing prompt embeddings, vector search, and telemetry logs.
* **The Management UI (React):** Administrative dashboard consuming the Flask metrics API.

---

## 2. Phase 1: Local LLM Setup (Ollama)

### Requirement 1.1: OpenAI-Compatible Stream Host
* **Objective:** Spin up a local inference server that mimics the exact JSON payload structures of cloud LLM providers.
* **Core Behaviors:**
    * Must accept standard POST requests matching the schema: `{"model": "llama3.2", "messages": [...], "stream": true}`.
    * Must output chunked text data utilizing Server-Sent Events (SSE).
* **References to Review:**
    * [Ollama GitHub Repository](https://github.com/ollama/ollama)
    * [Ollama OpenAI Compatibility Layer API Docs](https://github.com/ollama/ollama/blob/main/docs/api.md)

---

## 3. Phase 2: The Golang Data Plane (The Proxy Core)

### Requirement 2.1: Non-Buffering HTTP Reverse Proxy
* **Objective:** Intercept developer API calls on port `8080` and securely forward them to Ollama on port `11434`.
* **Core Behaviors:**
    * Transparently clone and forward HTTP headers, body streams, and verbs.
    * Strip custom gateway authentication tokens from incoming request headers before proxying upstream.
* **References to Review:**
    * [Go Standard Library: `httputil.ReverseProxy`](https://pkg.go.dev/net/http/httputil#ReverseProxy)

### Requirement 2.2: Immediate Stream Flushing
* **Objective:** Prevent the proxy from holding tokens in an internal buffer, ensuring immediate client delivery.
* **Core Behaviors:**
    * Intercept the chunked response from Ollama.
    * Cast the active HTTP response writer to a flushing interface to eject tokens chunk-by-chunk over the connection.
* **References to Review:**
    * [Go Standard Library: `http.Flusher` Interface](https://pkg.go.dev/net/http#Flusher)
    * [W3C Specification: Server-Sent Events (SSE)](https://html.spec.whatwg.org/multipage/server-sent-events.html)

### Requirement 2.3: Async Telemetry Offloading
* **Objective:** Track token usage and prompts without introducing latency into the client's network loop.
* **Core Behaviors:**
    * Upon client connection termination, count total processed stream tokens.
    * Instantly offload an asynchronous background network task to send this telemetry data to the Python Control Plane.
* **References to Review:**
    * [A Tour of Go: Goroutines](https://go.dev/tour/concurrency/1)

---

## 4. Phase 3: The Python Control Plane (The Intelligence Core)

### Requirement 3.1: Token/Prompt Embedding Pipeline
* **Objective:** Transform unstructured raw text strings into deterministic mathematical vectors.
* **Core Behaviors:**
    * Expose a Flask endpoint accepting a prompt string payload.
    * Generate a fixed-size embedding vector using a localized, lightweight semantic model.
* **References to Review:**
    * [SentenceTransformers Documentation](https://sbert.net/) (Look for `all-MiniLM-L6-v2` local inference)

### Requirement 3.2: Local Semantic Search Cache
* **Objective:** Intercept incoming prompts and evaluate if an identical or highly similar query was already evaluated.
* **Core Behaviors:**
    * Store vector footprints inside a lightweight local mathematical indexing tool.
    * Calculate vector distance metric (e.g., Cosine Similarity). If distance passes a **95% similarity ceiling**, return the cached response payload immediately.
* **References to Review:**
    * [FAISS (Facebook AI Similarity Search) Repository](https://github.com/facebookresearch/faiss)

### Requirement 3.3: Telemetry Database
* **Objective:** Log real-time cluster behavior and savings metrics securely.
* **Core Behaviors:**
    * Expose transactional REST endpoints for incoming Go metrics payloads.
    * Write incoming events (timestamp, raw tokens, prompt string, cache outcome) to a persistent SQLite database.
* **References to Review:**
    * [Flask-SQLAlchemy Documentation](https://flask-sqlalchemy.palletsprojects.com/)

---

## 5. Phase 4: The React Management UI (The Metrics Dashboard)

### Requirement 4.1: Operational Analytics View
* **Objective:** Provide administrative insight into cluster health, efficiency, and throughput.
* **Core Behaviors:**
    * Build a Single Page Application (SPA) designed to continuously pull data from the Flask control plane endpoints.
* **Component Visual Checklist:**
    * **Total Requests Ingested:** High-level scalar counter component.
    * **Cache Efficiency Ratio:** Metric highlighting hit-vs-miss percentages.
    * **Token Velocity Timeline:** Multi-series time chart visualizing token generation speeds and cache savings over time.
* **References to Review:**
    * [Recharts: React Charting Library](https://recharts.org/)
    * [SWR: React Hooks for Data Fetching](https://swr.vercel.app/)