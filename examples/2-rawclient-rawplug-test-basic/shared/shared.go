// Copyright (c) 2025 Michał Hodur
// SPDX-License-Identifier: MIT

package shared

import "github.com/mjwhodur/plugkit/codes"

type Ping struct {
}

type Pong struct {
	Message string
}

const (
	UnsupportedMessage codes.PluginExitReason = 8
)
