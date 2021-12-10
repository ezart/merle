// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"time"
)

const (
	MsgTypeCmd     = "cmd"
	MsgTypeCmdResp = "resp"
	MsgTypeSpam    = "spam"
)

type MsgType struct {
	Type string
	// payload
}

type MsgCmd struct {
	Type string
	Cmd  string
	// payload
}

type MsgCmdResp struct {
	Type string
	Cmd  string
	// payload
}

type MsgSpam struct {
	Type string
	Spam string
	// payload
}

const (
	CmdIdentify = "Identify"
	CmdStart    = "Start"
	CmdDevices  = "Devices"
)

type MsgIdentifyResp struct {
	Type        string
	Cmd         string
	Status      string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

type MsgDevicesDevice struct {
	Id     string
	Model  string
	Name   string
	Status string
}

type MsgDevicesResp struct {
	Type    string
	Cmd     string
	Devices []MsgDevicesDevice
}

type MsgStatusSpam struct {
	Type   string
	Spam   string
	Status string
	Id     string
	Model  string
	Name   string
}
