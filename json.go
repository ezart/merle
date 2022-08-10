// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build !tinygo
// +build !tinygo

package merle

import (
	"bytes"
	"encoding/json"
)

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func jsonPrettyPrint(msg []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, msg, "", "    "); err != nil {
		return ""
	}
	return prettyJSON.String() + "\n"
}
