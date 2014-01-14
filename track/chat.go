/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.13 (lee8oi)

chat and its methods are used to track current server chat messages.
*/

//
package track

import (
	"fmt"

	"html"
	"strings"
)

type message struct {
	Origin, Team, Type, Time, Text string
}

type chat struct {
	messages []message
}

func (c *chat) new(data string) {
	if len(data) > 1 {
		split := strings.Split(data, "\r")
		for _, value := range split {
			elem := strings.Split(strings.TrimSpace(value), "\t")
			if len(elem) < 6 {
				return
			}
			elem[5] = strings.Replace(html.EscapeString(elem[5]), `\`, `\\`, -1)
			m := message{
				Origin: elem[1],
				Team:   elem[2],
				Type:   elem[3],
				Time:   elem[4],
				Text:   elem[5],
			}
			c.messages = append(c.messages, m)
		}
	}
	return
}

//parse current chat messages.
func (c *chat) parse() {
	var base []message
	for key, value := range c.messages {
		fmt.Println(key, value)
	}
	c.messages = base
}