// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
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
let timerRunning = false
let refreshTimer

function sendForIdentity() {
	conn.send(JSON.stringify({Msg: "_GetIdentity"}))
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
		image.style.opacity = 1.0
	} else {
		image.style.opacity = 0.3
	}
	// force refresh of image by using getTime() trick
	image.src = "/" + id + "/assets/images/led-gpio17-" + on + ".png?t=" +
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
	if (timerRunning) {
		clearInterval(refreshTimer)
	}
	var date = new Date();
	setTimeout(function() {
		refreshTimer = setInterval(refreshUptime, 60000);
		timerRunning = true
		refreshUptime()
	}, (60 - date.getSeconds()) * 1000); // every minute, on the minute
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
		if (timerRunning) {
			clearInterval(refreshTimer)
		}
		location.reload(true)
	}

	conn.onmessage = function(evt) {
		var msg = JSON.parse(evt.data)

		console.log('blink', msg)

		switch(msg.Msg) {
		case "_SpamStatus":
			updateStatus(msg)
			break
		case "_ReplyIdentity":
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
		}
	}
}
