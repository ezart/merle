// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

// Thing base configuration
type BaseConfig struct {
	// [Optional] Thing's Id.  Ids are unique within an application
	// to differentiate one Thing from another.  Id is optional; if
	// Id is not given, a system-wide unique Id is assigned.
	Id string
	// Thing's Model.
	Model string
	// Thing's Name
	Name string
	// [Optional] system User.  If a User is given, any browser
	// views of the Thing's home page  will prompt for user/passwd.
	// HTTP Basic Authentication is used and the user/passwd given
	// must match the system creditials for the User.  If no User
	// is given, HTTP Basic Authentication is skipped; anyone can
	// view the home page.
	User string
	// [Optional] If PortPublic is given, an HTTP web server is
	// started on port PortPublic.  PortPublic is typically set to
	// 80.  The HTTP web server runs the Thing's home page.
	PortPublic uint
	// [Optional] If PortPublicTLS is given, an HTTPS web server is
	// started on port PortPublicTLS.  PortPublicTLS is typically
	// set to 443.  The HTTPS web server will self-certify using a
	// certificate from Let's Encrypt.  The public HTTPS server
	// will securely run the Thing's home page.  If PortPublicTLS
	// is given, PortPublic must be given.
	PortPublicTLS uint
	// [Optional] If PortPrivate is given, a HTTP server is
	// started on port PortPrivate.  This HTTP server does not
	// server up the Thing's home page but rather connects to
	// Thing's Mother using a websocket over HTTP.
	PortPrivate uint
	// [Optional] Run as Thing-prime.
	Prime bool
	// [Required, if Prime] PortPrime port is used to create a
	// tunnel from Thing to Thing-prime.  The port should be a
	// reserved port in ip_local_reserved_ports.
	PortPrime uint
}

// Thing configuration.  All Things share this configuration.
type ThingConfig struct {
	// [Optional] Thing's Id.  Ids are unique within an application
	// to differentiate one Thing from another.  Id is optional; if
	// Id is not given, a system-wide unique Id is assigned.
	Id string
	// Thing's Model.
	Model string
	// Thing's Name
	Name string
	// [Optional] system User.  If a User is given, any browser
	// views of the Thing's home page  will prompt for user/passwd.
	// HTTP Basic Authentication is used and the user/passwd given
	// must match the system creditials for the User.  If no User
	// is given, HTTP Basic Authentication is skipped; anyone can
	// view the home page.
	User string
	// [Optional] If PortPublic is given, an HTTP web server is
	// started on port PortPublic.  PortPublic is typically set to
	// 80.  The HTTP web server runs the Thing's home page.
	PortPublic uint
	// [Optional] If PortPublicTLS is given, an HTTPS web server is
	// started on port PortPublicTLS.  PortPublicTLS is typically
	// set to 443.  The HTTPS web server will self-certify using a
	// certificate from Let's Encrypt.  The public HTTPS server
	// will securely run the Thing's home page.  If PortPublicTLS
	// is given, PortPublic must be given.
	PortPublicTLS uint
	// [Optional] If PortPrivate is given, a HTTP server is
	// started on port PortPrivate.  This HTTP server does not
	// server up the Thing's home page but rather connects to
	// Thing's Mother using a websocket over HTTP.
	PortPrivate uint
	// [Optional] Run as Thing-prime.
	IsPrime bool
	// [Required, if Prime] PortPrime port is used to create a
	// tunnel from Thing to Thing-prime.  The port should be a
	// reserved port in ip_local_reserved_ports.
	PortPrime uint
	MaxConnections uint

	// This section describes a Thing's Mother.  Every Thing has a Mother.  A
	// Mother is also a Thing.  We can build a hierarchy of Things, with a Thing
	// having a Mother, a GrandMother, a Great GrandMother, etc.

	// Mother's Host address
	MotherHost string
	// User on Host
	MotherUser string
	// Port on Host for Mother's private HTTP server
	MotherPortPrivate uint

	// Bridge configuration.  A Thing implementing the Bridger interface will use
	// this config for bridge-specific configuration.

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
	Id: "",
	Model: "Thing",
	Name: "Thingy",
	User: "",
	PortPublic: 80,
	PortPublicTLS: 0,
	PortPrivate: 8080,
	IsPrime: false,
	PortPrime: 8000,
	MaxConnections: 10,
	MotherHost: "",
	MotherUser: "",
	MotherPortPrivate: 8080,
	BridgePortBegin: 8000,
	BridgePortEnd: 8040,
}
