package merle

import (
	"time"
)

const (
	CmdRun = "_CmdRun"
	CmdRunPrime = "_CmdRunPrime"
	GetIdentity = "_GetIdentity"
	ReplyIdentity = "_ReplyIdentity"
	SpamStatus = "_SpamStatus"
	GetChildren = "_GetChildren"
	ReplyChildren = "_ReplyChildren"
)

type MsgIdentity struct {
	Msg         string
	Status      string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

type MsgSpamStatus struct {
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
}

type MsgChild struct {
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
}

type MsgChildren struct {
	Msg      string
	Children []MsgChild
}
