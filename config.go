// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"flag"
)

// Thing configuration.  All Things share this configuration.
type ThingConfig struct {

	// The section describes a Thing.
	Thing struct {
		// [Optional] Thing's Id.  Ids are unique within an application
		// to differentiate one Thing from another.  Id is optional; if
		// Id is not given, a system-wide unique Id is assigned.
		Id string `yaml:"Id"`
		// Thing's Model.
		Model string `yaml:"Model"`
		// Thing's Name
		Name string `yaml:"Name"`
		// [Optional] system User.  If a User is given, any browser
		// views of the Thing's home page  will prompt for user/passwd.
		// HTTP Basic Authentication is used and the user/passwd given
		// must match the system creditials for the User.  If no User
		// is given, HTTP Basic Authentication is skipped; anyone can
		// view the home page.
		User string `yaml:"User"`
		// [Optional] If PortPublic is given, an HTTP web server is
		// started on port PortPublic.  PortPublic is typically set to
		// 80.  The HTTP web server runs the Thing's home page.
		PortPublic uint `yaml:"PortPublic"`
		// [Optional] If PortPublicTLS is given, an HTTPS web server is
		// started on port PortPublicTLS.  PortPublicTLS is typically
		// set to 443.  The HTTPS web server will self-certify using a
		// certificate from Let's Encrypt.  The public HTTPS server
		// will securely run the Thing's home page.  If PortPublicTLS
		// is given, PortPublic must be given.
		PortPublicTLS uint `yaml:"PortPublicTLS"`
		// [Optional] If PortPrivate is given, a HTTP server is
		// started on port PortPrivate.  This HTTP server does not
		// server up the Thing's home page but rather connects to
		// Thing's Mother using a websocket over HTTP.
		PortPrivate uint `yaml:"PortPrivate"`
		// [Optional] Run as Thing-prime.
		Prime bool `yaml:"Prime"`
		// [Required, if Prime] PortPrime port is used to create a
		// tunnel from Thing to Thing-prime.  The port should be a
		// reserved port in ip_local_reserved_ports.
		PortPrime uint `yaml:"PortPrime"`
	} `yaml:"Thing"`

	// [Optional] This section describes a Thing's Mother.  Every Thing has
	// a Mother.  A Mother is also a Thing.  We can build a hierarchy of
	// Things, with a Thing having a Mother, a GrandMother, a Great
	// GrandMother, etc.
	Mother struct {
		// Mother's Host address
		Host string `yaml:"Host"`
		// User on Host
		User string `yaml:"User"`
		// Port on Host for Mother's private HTTP server
		PortPrivate uint `yaml:"PortPrivate"`
	} `yaml:"Mother"`

	// [Optional] Bridge configuration.  A Thing implementing the Bridger
	// interface will use this config for bridge-specific configuration.
	Bridge struct {
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
		PortBegin uint `yaml:"PortBegin"`
		// Ending port number.
		PortEnd uint `yaml:"PortEnd"`
	} `yaml:"Bridge"`
}

func FlagThingConfig(id, model, name, user string) *ThingConfig {
	var cfg ThingConfig

	flag.BoolVar(&cfg.Thing.Prime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&cfg.Thing.PortPrime, "pport", 0, "Prime Port")

	flag.StringVar(&cfg.Thing.Id, "id", id, "Thing ID")
	flag.StringVar(&cfg.Thing.Model, "model", model, "Thing model")
	flag.StringVar(&cfg.Thing.Name, "name", name, "Thing name")

	flag.StringVar(&cfg.Thing.User, "luser", user,
		"Local user for HTTP Basic Authentication")
	flag.UintVar(&cfg.Thing.PortPublic, "lport", 80,
		"Local public HTTP listening port")
	flag.UintVar(&cfg.Thing.PortPublicTLS, "lportTLS", 0,
		"Local public HTTPS listening port (default 0, but usually 443)")
	flag.UintVar(&cfg.Thing.PortPrivate, "lportPriv", 8080,
		"Local private HTTP listening port")

	flag.StringVar(&cfg.Mother.Host, "rhost", "",
		"Remote host name or IP address")
	flag.StringVar(&cfg.Mother.User, "ruser", user,
		"Remote user")
	flag.UintVar(&cfg.Mother.PortPrivate, "rport", 8080,
		"Remote private HTTP listening port")

	return &cfg
}

func FlagBridgeConfig(id, model, name, user string, pbegin, pend uint) *ThingConfig {
	cfg := FlagThingConfig(id, model, name, user)

	flag.UintVar(&cfg.Bridge.PortBegin, "bbport", pbegin, "Bridge begin port")
	flag.UintVar(&cfg.Bridge.PortEnd, "beport", pend, "Bridge end port")

	return cfg
}

/*
func YamlThingConfig(file string) (*ThingConfig, error) {
	var cfg ThingConfig

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Opening config file failure: %s", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("Config decode error: %s", err)
	}

	return &cfg, nil
}
*/
