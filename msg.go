package merle

import (
	"time"
)

const (
	MsgTypeCmd         = "cmd"
	MsgTypeCmdResp     = "resp"
	MsgTypeSpam        = "spam"
)

type MsgType struct {
	Type	string
}

const (
	CmdIdentify = "Identify"
	CmdStart    = "Start"
)

type MsgIdentifyResp struct {
	Type        string
	Cmd         string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

type MsgCmd struct {
	Type	string
	Cmd	string
	// payload
}

type MsgCmdResp struct {
	Type	string
	Cmd	string
	// payload
}

type MsgSpam struct {
	Type	string
	Spam	string
	// payload
}
