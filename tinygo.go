// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build tinygo
// +build tinygo

package merle

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/wifinina"
)

type tunnel struct {
}

func newTunnel(t *Thing, host, user string,
	portPrivate, portRemote uint) *tunnel {
	return &tunnel{}
}

func (t *tunnel) start() {
}

func (t *tunnel) stop() {
}

type port struct {
}

type portAttachCb func(*port, *MsgIdentity) error

func newPort(thing *Thing, p uint, attachCb portAttachCb) *port {
	return &port{}
}

func (t *Thing) setAssetsDir(child *Thing) {
}

func (t *Thing) setHtmlTemplate() {
}

func (t *Thing) primeAttach(p *port, msg *MsgIdentity) error {
	return nil
}

func (t *Thing) primeRun() error {
	return nil
}

type Bridger interface {
}

type bridge struct {
}

func (b *bridge) getChild(id string) *Thing {
	return nil
}

func (b *bridge) start() {
}

func (b *bridge) stop() {
}

func newBridge(thing *Thing, portBegin, portEnd uint) *bridge {
	return &bridge{}
}

type web struct {
	public  *webPublic
	private *webPrivate
}

func newWeb(t *Thing, portPublic, portPublicTLS, portPrivate uint, user string) *web {
	return &web{}
}

func (w *web) handlePrimePortId() {
}

func (w *web) handleBridgePortId() {
}

func (w *web) staticFiles(t *Thing) {
}

type webPrivate struct {
}

func newWebPrivate(t *Thing, port uint) *webPrivate {
	return &webPrivate{}
}

func (w *webPrivate) start() {
}

func (w *webPrivate) stop() {
}

type webPublic struct {
}

func newWebPublic(t *Thing, port, portTLS uint, user string) *webPublic {
	return &webPublic{}
}

func (w *webPublic) start() {
}

func (w *webPublic) stop() {
}

type webSocket struct {
}

type wireSocket struct {
}

type logger struct {
}

func NewLogger(prefix string) *logger {
	return &logger{}
}

func (l *logger) printf(format string, v ...interface{}) {
}

func (l *logger) println(v ...interface{}) {
}

func (l *logger) fatalln(v ...interface{}) {
}

// TODO encoding/json isn't working with tinygo yet, so
// TODO these stubs for Marshal and Unmarshal need to be
// TODO open-coded to work on basic Merle message types
// TODO like CmdInit and CmdRun and whatever else we
// TODO might run into on a  tinygo deployment.

func jsonMarshal(v interface{}) ([]byte, error) {
	return []byte{}, nil
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return nil
}

func jsonPrettyPrint(msg []byte) string {
	return ""
}

func Nano33ConnectAP(ssid, pass string) {
	// These are the default pins for the Arduino Nano33 IoT.
	spi := machine.NINA_SPI

	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	// This is the ESP chip that has the WIFININA firmware flashed on it
	adaptor := wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	adaptor.Configure()

	// Connect to access point
	time.Sleep(2 * time.Second)
	println("Connecting to " + ssid)
	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil { // error connecting to AP
		for {
			println(err)
			time.Sleep(1 * time.Second)
		}
	}

	println("Connected.")

	time.Sleep(2 * time.Second)
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		println(err.Error())
		time.Sleep(1 * time.Second)
	}
	println(ip.String())
}
