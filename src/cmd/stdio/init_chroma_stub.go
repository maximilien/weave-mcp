//go:build !darwin || !cgo

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package main

// Chroma is not available on this platform (requires macOS + CGO)
// The weave-cli stub will be registered automatically via its init() function
