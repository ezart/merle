// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import "time"

// System messages.  System messages are prefixed with '_'.
const (
	// CmdInit is guaranteed to be the first message a new Thing will see.
	// Thing can optionally subscribe and handle CmdInit via Subscribers(),
	// to initialize Thing's state.
	//
	// CmdInit is not sent to Thing Prime.  Thing Prime will get its
	// initial state with a GetState call to Thing.
	CmdInit = "_CmdInit"

	// CmdRun is Thing's main loop.  All Things must subscribe and handle
	// CmdRun, via Subscribers().  CmdRun should run forever; it is an error
	// for CmdRun handler to exit.
	//
	// CmdRun is not sent to Thing Prime.  Thing Prime does not have a main
	// loop.
	//
	// If the Thing is a bridge, CmdRun is also sent to the bridge bus on
	// startup of the bridge, via BridgeSubscribers().  In this case, CmdRun
	// is optional and doesn't need to run forever.
	CmdRun = "_CmdRun"

	// GetIdentity requests Thing's identity.  Thing does not need to
	// subscribe to GetIdentity.  Thing will internally respond with a
	// ReplyIdentity message.
	GetIdentity = "_GetIdentity"

	// Response to GetIdentity.  ReplyIdentity message is coded as
	// MsgIdentity.
	ReplyIdentity = "_ReplyIdentity"

	// GetState requests Thing's state.  Thing should respond with a
	// ReplyState message containing Thing's state.
	GetState = "_GetState"

	// Response to GetState.  ReplyState message coding is Thing-specific.
	//
	// It is convienent to use Thing's type struct (the Thinger) as the
	// container for Thing's state.  Just include a Msg member and export
	// any other state members (with an uppercase leading letter).  Then
	// the whole type struct can be passed in p.Marshal() to form the
	// response.
	//
	//	type thing struct {
	//		Msg       string
	//		StateVar0 int
	//		StateVar1 bool
	//		// non-exported members
	//	}
	//
	//	func (t *thing) init(p *merle.Packet) {
	//		t.StateVar0 = 42
	//		t.StateVar1 = true
	//	}
	//
	//	func (t *thing) getState(p *merle.Packet) {
	//		t.Msg = merle.ReplyState
	//		p.Marshal(t).Reply()
	//	}
	//
	// Will send JSON message:
	//
	//  {
	//	"Msg": "_ReplyState",
	//	"StateVar0": 42,
	//	"StateVar1": true,
	//  }
	ReplyState = "_ReplyState"

	// EventStatus message is an unsolicited notification that a child
	// Thing's connection status has changed.
	//
	// EventStatus message is coded as MsgEventStatus.
	EventStatus = "_EventStatus"
)

// All messages in Merle build on this basic struct.  All messages have a
// member Msg which is the message type, a string that's unique within the
// Thing's message namespace.
//
// System messages type Msg is prefixed with a "_".  Regular Thing messages
// should not be prefixed with "_".
type Msg struct {
	Msg string
	// Message-specific members here
}

// Event status change notification message.  On child connect or disconnect,
// this notification is sent to:
//
// 1. If Thing Prime, send to all listeners (browsers) on Thing Prime.
// 2. If Bridge, send to mother bus and to bridge bus.
type MsgEventStatus struct {
	Msg    string
	Id     string
	Online bool
}

// Thing identification message return in ReplyIdentity
type MsgIdentity struct {
	Msg         string
	Id          string
	Model       string
	Name        string
	Online      bool
	StartupTime time.Time
}
