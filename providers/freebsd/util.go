// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package freebsd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
)

func parseKeyValue(content []byte, separator string, callback func(key, value []byte) error) error {
	sc := bufio.NewScanner(bytes.NewReader(content))
	for sc.Scan() {
		parts := bytes.SplitN(sc.Bytes(), []byte(separator), 2)
		if len(parts) != 2 {
			continue
		}

		if err := callback(parts[0], bytes.TrimSpace(parts[1])); err != nil {
			return err
		}
	}

	return sc.Err()
}

func findValue(filename, separator, key string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var line []byte
	sc := bufio.NewScanner(bytes.NewReader(content))
	for sc.Scan() {
		if bytes.HasPrefix(sc.Bytes(), []byte(key)) {
			line = sc.Bytes()
			break
		}
	}
	if len(line) == 0 {
		return "", fmt.Errorf("%v not found", key)
	}

	parts := bytes.SplitN(line, []byte(separator), 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected line format for '%v'", string(line))
	}

	return string(bytes.TrimSpace(parts[1])), nil
}

func decodeBitMap(s string, lookupName func(int) string) ([]string, error) {
	mask, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return nil, err
	}

	var names []string
	for i := 0; i < 64; i++ {
		bit := mask & (1 << uint(i))
		if bit > 0 {
			names = append(names, lookupName(i))
		}
	}

	return names, nil
}

func parseBytesOrNumber(data []byte) (uint64, error) {
	parts := bytes.Fields(data)

	if len(parts) == 0 {
		return 0, errors.New("empty value")
	}

	num, err := strconv.ParseUint(string(parts[0]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value: %w", err)
	}

	var multiplier uint64 = 1
	if len(parts) >= 2 {
		switch string(parts[1]) {
		case "kB":
			multiplier = 1024
		default:
			return 0, fmt.Errorf("unhandled unit %v", string(parts[1]))
		}
	}

	return num * multiplier, nil
}
