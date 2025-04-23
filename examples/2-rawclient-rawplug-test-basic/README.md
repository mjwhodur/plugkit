# 🔌 SmartPlug Test Scenario

This is the **simplest working SmartPlug** based on
the [PlugKit](https://github.com/...) protocol — a minimal but functional
demonstration of plugin-to-host communication.

## 🧠 Concept

The goal of this test scenario is to verify that a single message can be sent
from a host to a plug, and that the plug can respond correctly — exactly once.
No loops, no protocol magic. Just one roundtrip. Clean and minimal.

This is useful as a smoke test for your PlugKit-compatible infrastructure or as
a starting point for implementing more complex plugins.

## 📁 Project structure

```
. 
├── plug/ # 🔌 The plug binary – implements RawStreamPlug 
├── client/ # 🧑‍💻 The host – sends a single message to the plug 
└── shared/ # 📦 Shared types and helpers – messages, CBOR encoding, codes, etc.
```

- The **plug** listens for a single `Envelope` message, processes it, and
  responds.
- The **host** sends one message, waits for the reply, and exits.
- The **shared** package contains the agreed contract between plug and host (
  message types, payload encoding, etc).

## ✅ What it demonstrates

- 🎯 PlugKit-compatible message envelope handling (CBOR-encoded)
- 🔄 One-shot request/response cycle
- 🔌 Bidirectional stdio stream between host and plug

## 🚀 Running the scenario

You can run the test by building both binaries.

The host will:

1. Start the plug process

1. Send a single CBOR-encoded message

1. Wait for the plug to respond

1. Print the response and exit

The plug will:

1. Wait for a message

1. Process it

1. Reply

1. Shut down

This is a great minimal example to understand the lifecycle of a SmartPlug 🔌 –
message in, response out, context-aware shutdown, and no extra fluff.

Enjoy plugging!