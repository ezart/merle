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

	EventStatus = "_EventStatus"
)

// Basic message structure.  All messages in Merle build on this basic struct,
// with a member Msg which is the message type, something unique within the
// Thing's message namespace.
type Msg struct {
	Msg string
	// Message-specific members here
}

// Event status change notification message.  On child connect or disconnect,
// this notification is sent to:
//
// 1) If Thing Prime, send to all listeners (browsers) on Thing Prime.
// 2) If Bridge, send to mother bus and to bridge bus.
type MsgEventStatus struct {
	Msg    string
	Id     string
	Online bool
}

// ReplyIdentity returns MsgIdentity response
type MsgIdentity struct {
	Msg         string
	Id          string
	Model       string
	Name        string
	Online      bool
	StartupTime time.Time
}
