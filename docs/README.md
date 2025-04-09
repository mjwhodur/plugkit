# PlugKit Docs


**PlugKit** to stupidly simple, stream-capable plugin runtime for Go.  
No gRPC. No protobuf. No codegen.  
Just structs, CBOR, pipes and a handshake.

## ✨ Why PlugKit?

PlugKit is a micro-framework that lets you:

- create and run plugins as separate processes (we call them simply plugs)
- communicate with them over `stdin`/`stdout`
- pass arbitrary Go structs encoded with CBOR
- use bidirectional message routing
- gracefully terminate a plug whenever you want (`Finish`)

It's like HashiCorp’s `go-plugin`, but:
- without the pain
- without reflection
- without type registries
- **with streaming message support** 😎 (in future)

