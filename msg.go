// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"time"
)

// System messages.  System messages are prefixed with '_'.
const (
	// All Things must handle CmdRun to do work.  CmdRun should run
	// forever; it is an error for CmdRun handler to finish.  The callback
	// merle.RunForever can be used if there is no more work beyond select{}.
	// CmdRun is not sent to Thing-prime...see CmdRunPrime.
	CmdRun = "_CmdRun"
	// CmdRunPrime is sent to Thing-prime rather than CmdRun.  Unlike
	// CmdRun, CmdRunPrime should not run forever.  CmdRunPrime lets
	// Thing-prime fetch it's state from Thing.
	CmdRunPrime = "_CmdRunPrime"
	// GetIdentity requests Thing's identity
	GetIdentity = "_GetIdentity"
	// Reply to GetIdentity with MsgIdentity
	ReplyIdentity = "_ReplyIdentity"
	// SpamStatus is sent when Thing's status (online, offline, etc)
	// changes
	SpamStatus = "_SpamStatus"
	// GetChildren requests Bridge's children
	GetChildren = "_GetChildren"
	// Reply to GetChildren with MsgChildren
	ReplyChildren = "_ReplyChildren"
)

// ReplyIdentity returns MsgIdentity response
type MsgIdentity struct {
	Msg         string
	Status      string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

// MsgSpamStatus is sent when Thing's status (online, offline, etc) changes.
// Listeners can update their status and forward with p.Broadcast().
type MsgSpamStatus struct {
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
}

// A Bridge child
type MsgChild struct {
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
}

// ReplyChildren returns MsgChildren, a list of the Bridge's children
// (MsgChild)
type MsgChildren struct {
	Msg      string
	Children []MsgChild
}
