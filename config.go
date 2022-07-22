// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

// Thing configuration.  A default configuration is assigned at creation
// (NewThing()).  Override default configurations before calling thing.Run().
// For example:
//
// func main() {
// 	thing := merle.NewThing(&hello{})
// 	thing.Cfg.User = "merle"  // turn on Basic Authentication for user merle
// 	thing.Cfg.PortPublic = 80 // turn on public web server on port :80
// 	log.Fatalln(thing.Run())
// }

type ThingConfig struct {

	// ########## Thing configuration.
	//
	// [Optional] Thing's Id.  Ids are unique within an application to
	// differentiate one Thing from another.  Id is optional; if Id is not
	// given, a system-wide unique Id is assigned.
	Id string

	// Thing's Model.  The default is "Thing".
	Model string

	// Thing's Name.  The default is "Thingy".
	Name string

	// [Optional] system User.  If a User is given, any browser views of
	// the Thing's UI will prompt for user/passwd.  HTTP Basic
	// Authentication is used and the user/passwd given must match the
	// system creditials for the User.  If no User is given, HTTP Basic
	// Authentication is skipped; anyone can view the UI.  The default is
	// "".
	User string

	// [Optional] If PortPublic is non-zero, an HTTP web server is started
	// on port PortPublic.  PortPublic is typically set to 80.  The HTTP
	// web server runs Thing's UI.  The default is 0.
	PortPublic uint

	// [Optional] If PortPublicTLS is non-zero, an HTTPS web server is
	// started on port PortPublicTLS.  PortPublicTLS is typically set to
	// 443.  The HTTPS web server will self-certify using a certificate
	// from Let's Encrypt.  The public HTTPS server will securely run the
	// Thing's UI.  If PortPublicTLS is given, PortPublic must be given.
	// The default is 0.
	PortPublicTLS uint

	// [Optional] If PortPrivate is non-zero, a private HTTP server is
	// started on port PortPrivate.  This HTTP server does not server up
	// the Thing's UI but rather connects to Thing's Mother using a
	// websocket over HTTP.  The default is 0.
	PortPrivate uint

	// [Optional] Run as Thing-prime.  The default is false.
	IsPrime bool

	// [Required, if Prime] PortPrime port is used to create a
	// tunnel from Thing to Thing-prime.  The port should be a
	// reserved port in ip_local_reserved_ports.
	PortPrime uint

	// MaxConnection is maximum number of inbound connections to a Thing.
	// Inbound connections are WebSockets from web browsers or WebSockets
	// from Thing Prime.  The default is 30.  With the default, the 31st
	// (and higher) concurrent WebSocket connection attempt will block,
	// waiting for one of the first 30 WebSocket sessions to terminate.
	MaxConnections uint

	// ########## Mother configuration.
	//
	// This section describes a Thing's mother.  Every Thing has a mother.  A
	// mother is also a Thing.  We can build a hierarchy of Things, with a Thing
	// having a mother, a grandmother, a great grandmother, etc.
	//
	// Mother's Host address.  This the IP address or Domain Name of the
	// host running mother.  Host address can be on the local network or across
	// the internet.
	MotherHost string

	// User on host with SSH access into host.  Host should be configured
	// with user's public key so SSH access is password-less.
	MotherUser string

	// Port on Host for Mother's private HTTP server
	MotherPortPrivate uint

	// ########## Bridge configuration.
	//
	// A Thing implementing the Bridger interface will use this config for
	// bridge-specific configuration.
	//
	// Beginning port number.  The bridge will listen for Thing
	// (child) connections on the port range [BeginPort-EndPort].
	//
	// The bridge port range must be within the system's
	// ip_local_reserved_ports.
	//
	// Set a range using:
	//
	//   sudo sysctl -w net.ipv4.ip_local_reserved_ports="8000-8040"
	//
	// Or, to persist setting on next boot, add to /etc/sysctl.conf:
	//
	//   net.ipv4.ip_local_reserved_ports = 8000-8040
	//
	// And then run sudo sysctl -p
	//
	BridgePortBegin uint

	// Ending port number.
	BridgePortEnd uint
}

var defaultCfg = ThingConfig{
	Id:                "",
	Model:             "Thing",
	Name:              "Thingy",
	User:              "",
	PortPublic:        0,
	PortPublicTLS:     0,
	PortPrivate:       0,
	IsPrime:           false,
	PortPrime:         8000,
	MaxConnections:    30,
	MotherHost:        "",
	MotherUser:        "",
	MotherPortPrivate: 8080,
	BridgePortBegin:   8000,
	BridgePortEnd:     8040,
}
