// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let conn

let id
let model
let name
let status
let ledState
let paused = false
let refreshTimer

function sendForIdentity() {
	conn.send(JSON.stringify({Msg: "GetIdentity"}))
}

function saveIdentity(msg) {
	id = msg.Id
	model = msg.Model
	name = msg.Name
	status = msg.Status
}

function refreshBase() {
	var labels = document.getElementById("lables")
	var preId = document.getElementById("id")
	var preModel = document.getElementById("model")
	var preName = document.getElementById("name")

	preId.textContent = id
	preModel.textContent = model
	preName.textContent = name

	labels.className = "labels"
}

function refreshAll() {
	refreshBase()
}

function updateStatus(msg) {
	status = msg.Status
	refreshAll()
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForIdentity()
	}

	conn.onclose = function(evt) {
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('event', msg)

		switch(msg.Msg) {
		case "ReplyIdentity":
			saveIdentity(msg)
			refreshAll()
			break
		}
	}
}
