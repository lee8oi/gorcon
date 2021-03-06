/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/log (lee8oi)

Source code for this package is based on the code by Gary Burd found at
http://gary.beagledreams.com/page/go-websocket-chat.html

*/

//
package log

import (
	"time"
)

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var H = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		}
	}
}

//log is an attempt to broadcast a log message to currently open sockets
func (h *hub) Log(s string) {
	h.broadcast <- []byte(s)
}

//loop is used for development/testing. Simply broadcasts the current date/time.
func (h *hub) Loop() {
	for {
		if h.broadcast != nil {
			h.Log(time.Now().String())
			time.Sleep(1 * time.Second)
		}
	}
}
