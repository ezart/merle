// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let conn

let id
let model
let name
let status
let startupTime
let ledState = false
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
	startupTime = new Date(msg.StartupTime)
}

function sendForPaused() {
	conn.send(JSON.stringify({Msg: "GetPaused"}))
}

function savePaused(msg) {
	paused = msg.Paused
	ledState = msg.State
}

function saveLedState(msg) {
	ledState = msg.State
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

function refreshLed() {
	var image = document.getElementById("raspi")
	on = "off"
	if (status == "online") {
		if (ledState) {
			on = "on"
		}
	} else {
		image.style.opacity = 0.3
	}
	// force refresh of image by using getTime() trick
	image.src = "./web/images/" + model + "/led-gpio17-" + on + ".png?t=" +
		new Date().getTime()
	image.style.visibility = "visible"
}

function refreshButton() {
	var button = document.getElementById("pause")
	if (paused && status == "online") {
		button.textContent = "Resume"
	} else {
		button.textContent = "Pause"
	}
	button.style.visibility = "visible"
	button.disabled = (status != "online")
}

function refreshUptime() {
	var uptime = document.getElementById("uptime")

	if (status != "online") {
		uptime.textContent = "<offline>"
		return
	}

	var diffMs = Math.abs(new Date() - startupTime)
	var diffMins = Math.floor(diffMs / 1000 / 60)
	var days = Math.floor(diffMins / 60 / 24)
	var hours = Math.floor((diffMins - (days * 24 * 60)) / 60)
	var minutes = Math.floor(diffMins - (hours * 60) - (days * 24 * 60))

	uptime.textContent = days + " days " + hours + " hours " + minutes + " mins"
}

function refreshAll() {
	refreshBase()
	refreshLed()
	refreshButton()
	refreshUptime()
	if (typeof refreshTimer == 'undefined') {
		refreshTimer = setInterval(refreshUptime, 1000 * 60)  // once a sec
	}
}

function updateStatus(msg) {
	status = msg.Status
	refreshAll()
}

function pause() {
	if (paused) {
		conn.send(JSON.stringify({Msg: "CmdResume"}))
	} else {
		conn.send(JSON.stringify({Msg: "CmdPause"}))
	}
	paused = !paused
	refreshButton()
}

function Run(scheme, host, id) {

	conn = new WebSocket(scheme + host + "/ws/" + id)

	conn.onopen = function(evt) {
		sendForIdentity()
	}

	conn.onclose = function(evt) {
		clearInterval(refreshTimer)
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('thing msg', msg)

		switch(msg.Msg) {
		case "ReplyIdentity":
			saveIdentity(msg)
			sendForPaused()
			break
		case "ReplyPaused":
			savePaused(msg)
			refreshAll()
			break
		case "SpamLedState":
			saveLedState(msg)
			refreshLed()
			break
		case "CmdPause":
			paused = true
			refreshButton()
			break
		case "CmdResume":
			paused = false
			refreshButton()
			break
		case "SpamStatus":
			updateStatus(msg)
			break
		}
	}
}
