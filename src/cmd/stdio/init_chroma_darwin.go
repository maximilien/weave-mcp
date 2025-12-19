//go:build darwin && cgo

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package main

import (
	// Chroma is only supported on macOS with CGO enabled
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)
