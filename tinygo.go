// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build tinygo
// +build tinygo

package merle

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
