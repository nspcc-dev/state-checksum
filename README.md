# Overview
This repository contains a PoC of key-value DB checksums based on hashes
XORing. It's main purpose is to show that using this approach it's possible to
reproduce overall DB state sum from DB changes.

# Usage and features

Requires Go 1.12+. Usage:
```
make test
make cover
```

The most interesting part is `TestCachedStateSimple` function that tests
various scenarios of cache and persistence store interactions. `Checksum`
implementation in `MemoryStore` always computes full checksum of the database
while `Checksum` in the MemCachedStore is computed incrementally during `Put`
and `Delete`.

# Implementation details
This codebase is based on neo-go repository (`pkg/core/storage`), so it
contains some useless (from a PoC point of view) code, lacks proper locking in
many places and doesn't have proper PutBatch implementation for
MemCachedStore. All of this is just because it's a quick proof of concept, so
don't expect it to be polished.
