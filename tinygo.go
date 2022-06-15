// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build tinygo
// +build tinygo

package merle

type ThingAssets struct {
	Dir          string
	Template     string
	TemplateText string
}

type web struct {
}

func newWeb(t *Thing, portPublic, portPublicTLS, portPrivate uint, user string) *web {
	return &web{}
}

func (w *web) start() {
}

func (w *web) stop() {
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

func (w *webPrivate) handlePrimePortId() {
}

func (w *webPrivate) handleBridgePortId() {
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

func (t *Thing) runOnPort(p *port) error {
	return nil
}
