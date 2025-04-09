# PlugKit

**PlugKit** to stupidly simple, stream-capable plugin runtime for Go.  
No gRPC. No protobuf. No codegen.  
Just structs, CBOR, pipes and a handshake.

## ✨ Why PlugKit?

PlugKit is a micro-framework that lets you:

- run plugins as separate processes
- communicate with them over `stdin`/`stdout`
- pass arbitrary Go structs encoded with CBOR
- use bidirectional message routing (the plugin can speak first!)
- gracefully terminate a plugin whenever you want (`Finish`)

It's like HashiCorp’s `go-plugin`, but:
- without the pain
- without reflection
- without type registries
- **with streaming message support** 😎 (in future)

## 🧪 Status

PlugKit is under active development.   
The API will evolve slightly, but the core already works:

- ✅ Running plugins as separate processes
- ✅ Bidirectional communication (host <-> plugin)
- ✅ CBOR serialization (`fxamacker/cbor`)
- ✅ Handling multiple message types
- ✅ `Finish()` with exit code support
- ⏳ Handshake with capabilities negotiation
- ⏳ Unit tests
- ⏳ API documentation  

