/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/track (lee8oi)

chat and its methods are used to track current server chat messages.
*/

//
package track

import (
	"fmt"
	"strings"
)

type message struct {
	Pid, Origin, Team, Type, Time, Text string
}

func parseChat(data string, com chan *message) {
	if len(data) > 1 {
		split := strings.Split(data, "\r")
		for _, value := range split {
			elem := strings.Split(strings.TrimSpace(value), "\t")
			if len(elem) < 5 {
				continue
			} else if len(elem) < 6 {
				elem = append(elem, " ")
			} else {
				//elem[5] = html.EscapeString(elem[5])
			}
			m := message{
				Pid:    elem[0],
				Origin: elem[1],
				Team:   elem[2],
				Type:   elem[3],
				Time:   elem[4],
				Text:   elem[5],
			}
			if len(m.Text) > 0 && strings.IndexAny(m.Text, "!/|") == 0 {
				com <- &m
			}
			fmt.Println(m)
		}
	}
	close(com)
}
