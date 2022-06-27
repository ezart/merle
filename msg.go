// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"time"
)

// System messages.  System messages are prefixed with '_'.
const (
	CmdInit = "_CmdInit"
	// All Things must handle CmdRun to do work.  CmdRun should run
	// forever; it is an error for CmdRun handler to finish.  The callback
	// merle.RunForever can be used if there is no more work beyond select{}.
	// CmdRun is not sent to Thing-prime.
	CmdRun = "_CmdRun"
	// GetIdentity requests Thing's identity
	GetIdentity = "_GetIdentity"
	// Reply to GetIdentity with MsgIdentity
	ReplyIdentity = "_ReplyIdentity"
	// GetState TODO doc
	GetState   = "_GetState"
	ReplyState = "_ReplyState"

	EventConnect = "_EventConnect"
	EventDisconnect = "_EventDisconnect"
	EventBridgeConnect = "_EventBridgeConnect"
	EventBridgeDisconnect = "_EventBridgeDisconnect"
)

type Msg struct {
	Msg string
}

// ReplyIdentity returns MsgIdentity response
type MsgIdentity struct {
	Msg         string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}
