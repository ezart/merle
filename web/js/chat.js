// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

let conn

let id
let model
let name
let status
let user = ""

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

function sendNewUser(id) {
	conn.send(JSON.stringify({Msg: "CmdNewUser", User: id}))
}

function newUser(id) {
	var users = document.getElementById("users")
	var newpre = document.createElement("pre")

	user = id
	newpre.innerText = id 
	users.appendChild(newpre)
}

function sendText(text) {
	conn.send(JSON.stringify({Msg: "CmdText", Text: text}))
}

function newText(text) {
	var newpre = document.createElement("pre")
	var history = document.getElementById("history")

	if (user != "") {
		newpre.innerText = text
		history.appendChild(newpre)
	}
}

function enter(event) {
	if (event.keyCode == 13)
		document.getElementById('send').click()
}

function send() {
	var button = document.getElementById("send")
	var input = document.getElementById("input")

	if (input.value == "") {
		return
	}

	if (button.textContent == "Login") {
		button.textContent = "Send"
		newUser(input.value)
		sendNewUser(input.value)
		return
	}

	text = input.value
	sendText(text)
	newText(text)
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

		console.log('thing msg', msg)

		switch(msg.Msg) {
		case "ReplyIdentity":
			saveIdentity(msg)
			refreshAll()
			break
		case "CmdNewUser":
			newUser(msg.User)
			break
		case "CmdText":
			newText(msg.Text)
			break
		}
	}
}
