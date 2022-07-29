// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build !tinygo
// +build !tinygo

package merle

import (
	"log"
	"os"
)

type logger struct {
	log     *log.Logger
	enabled bool
}

func newLogger(prefix string, enabled bool) *logger {
	return &logger{log: log.New(os.Stderr, prefix, 0), enabled: enabled}
}

func (l *logger) printf(format string, v ...interface{}) {
	if l.enabled {
		l.log.Printf(format, v...)
	}
}

func (l *logger) println(v ...interface{}) {
	if l.enabled {
		l.log.Println(v...)
	}
}

func (l *logger) fatalln(v ...interface{}) {
	if l.enabled {
		l.log.Fatalln(v...)
	}
}
