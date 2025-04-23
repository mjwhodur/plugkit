# Examples

This folder has working examples that are compiled during testing of the library
These examples are integration tests of the PlugKit Framework.
## Core integration tests 

### 0-plugin-test
It is a basic test for a SmartPlug and SmartClient.
Checks whether marshalling is correct for Correct messages.
It is a very basic architecture - when your plugins are predictable.
Or just for fast prototyping.

### 1-rawclient-smartplug-test-basic
It is a basic test for a SmartPlug and RawClient.
Checks whether marshalling is correct for Correct messages.
It is a very basic architecture - when your plugins are predictable and correctly written.

### 2-rawclient-rawplug-test-basic
It is a basic test for a RawPlug and RawClient.
Checks whether marshalling is correct for Correct messages.
It is a very basic architecture - when your plugins are predictable and correctly written.

# ✅ PlugKit – Integration Test Checklist

Each combination of plugin and client implementation should pass the following scenarios to be considered conformant with the PlugKit protocol.

---

## 🧪 Core Scenarios

- [ ] **Basic successful exchange**
    - The client sends a valid request, and the plugin responds with a correct `PluginResponse`.

- [ ] **Unknown message type**
    - The client sends an unsupported message → the plugin should return `Unsupported`.

- [ ] **Plugin-side operational error**
    - Plugin returns a `PluginResponse`, but `Handle()` returns an error → should be interpreted as `HandlingError`.

- [ ] **Plugin crash (panic or process exit)**
    - The client should handle this as `PlugCrashed` or similar error.

- [ ] **Graceful plugin completion**
    - Plugin sends a `PluginFinish` with `OperationSuccess`.

- [ ] **Plugin ends with error**
    - Plugin sends a `PluginFinish` with `OperationError` and an error message.

- [ ] **Plugin responds with a known message, but client has no handler**
    - Client should return `Unsupported`.

- [ ] **Handler returns an error**
    - Should result in a `HandlingError`.

---

## 🧪 Sanity and edge-case checks

- [ ] **Envelope with unknown protocol version**
    - `Version: 99` → client should ignore or reject it gracefully.

- [ ] **Envelope missing a `Type`**
    - `Type: ""` → client should treat this as `Unsupported` or invalid.

- [ ] **Valid `Type`, but malformed `Raw` (valid CBOR, wrong structure)**
    - `Raw` decodes but doesn't match expected structure → client should return `PluginToHostCommunicationError`.

- [ ] **`PluginResponse` with missing `Result.Type`**
    - Empty `Result` structure → client should handle or reject gracefully.

- [ ] **Empty byte array in `Raw`**
    - `Raw: []byte{}` instead of `nil` → check how unmarshalling behaves.

---

## 🚫 Protocol violation and resilience

- [ ] **Malformed CBOR (non-CBOR, JSON, text)**
    - Plugin sends invalid binary → client returns `PayloadMalformed`.

- [ ] **Plugin closes pipe without sending `FinishMessage`**
    - Client encounters `io.EOF` → should return `PlugCrashed`.

- [ ] **Plugin sends `FinishMessage`, then more data**
    - Client should ignore or reject any messages after completion.

- [ ] **Plugin sends two `PluginResponse`s without flushing**
    - Buffered data – client should still decode and handle both.

---

## ⛔️ Unexpected or misused behavior

- [ ] **RunCommand before Start**
    - Calling `RunCommand()` without `Start()` should return `PlugNotStarted`.

- [ ] **Invalid plugin binary path**
    - Plugin executable not found → client returns startup error (`exec: not found`, etc.).

- [ ] **Plugin echoes `Unsupported` back repeatedly**
    - Infinite error loop → client should prevent endless message ping-pong.

---

## 📌 Optional: Streaming mode (`RawStreamClient` / `RawStreamPlug`)

- [ ] **Multiple consecutive messages**
    - Plugin sends 2–3 messages in a row → client processes each correctly.

- [ ] **Graceful shutdown from client**
    - Calling `Stop()` on the stream client cancels `Run()` and invokes `CloseSignal()` on the plugin.

---

## 📝 Notes

- Tests may be implemented using `go test`, executable `examples/`, shell scripts, or other approaches — the key is that they run automatically.
- Examples provided with the framework must pass these tests before any release.
