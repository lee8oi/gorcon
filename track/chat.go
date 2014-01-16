/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track version 14.1.15 (lee8oi)

chat and its methods are used to track current server chat messages.
*/

//
package track

import (
	"fmt"
	//"html"
	"strings"
)

type message struct {
	Origin, Team, Type, Time, Text string
}

type chat struct {
	messages []message
}

//add takes 'bf2cc clientchatbuffer' string and appends all messages to chat.
func (c *chat) add(data string) {
	if len(data) > 1 {
		split := strings.Split(data, "\r")
		for _, value := range split {
			elem := strings.Split(strings.TrimSpace(value), "\t")
			if len(elem) < 5 {
				continue
			} else if len(elem) < 6 {
				elem = append(elem, " ")
			} else {
				//elem[5] = strings.Replace(html.EscapeString(elem[5]), `\`, `\\`, -1)
			}
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

//parse existing chat messages. Checks for command prefixes in message text and
//sends the lines to com channel.
func (c *chat) parse(pl *playerList, com chan *player) {
	for _, value := range c.messages {
		cmd := c.check(value)

		if len(cmd) > 0 {
			p := pl.player(value.Origin)
			p.Command = cmd
			com <- p
		}
		fmt.Println(value)
	}
	close(com)
	c.clear()
}

//check message for command prefixes then return command line.
func (c *chat) check(value message) (cmd string) {
	if len(value.Text) > 1 {
		trimd := strings.TrimSpace(strings.TrimPrefix(value.Text, ": ")) //for testing commands via admin chat
		if strings.IndexAny(trimd, "!/|") == 0 {
			cmd = trimd[1:]
		}
	}
	return
}

//clear all chat messages
func (c *chat) clear() {
	var base []message
	c.messages = base
}
