/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon/log (lee8oi)

Source code for this package is based on the code by Gary Burd found at
http://gary.beagledreams.com/page/go-websocket-chat.html

*/

/*
The log package is used to log messages to the web via gorilla websocket. This
version of the package is intended to be used with gorcon.
*/
package log

import (
	"flag"
	"log"
	"net/http"
	"text/template"
)

var addr = flag.String("addr", ":23456", "http service address")
var homeTempl = template.Must(template.ParseFiles("home.html"))

func homeHandler(c http.ResponseWriter, req *http.Request) {
	homeTempl.Execute(c, req.Host)
}

//func main() {
//	//flag.Parse()

//}

func Start() {
	go H.run()
	//go H.Loop()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
