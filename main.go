/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

gorcon version 14.1.13 (lee8oi)

gorcon package contains the essential functions needed for, connecting to &
running commands on, BF2CC based Rcon servers.

*/
package gorcon

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"net"
	"strings"
	"time"
)

type Config struct {
	Admin, Address, Port, Pass string
}

type Rcon struct {
	admin, pass, seed, status, service string
	reconnect                          bool
	wait                               time.Duration
	sock                               net.Conn
	send                               chan []byte
	queue                              chan string
	receive                            chan string
}

//AutoReconnect enables reconnection. Valid time units are "ns", "us" (or "µs"),
// "ms", "s", "m", "h" (see time.ParseDuration doc).
func (r *Rcon) AutoReconnect(wait string) {
	dur, err := time.ParseDuration(wait)
	if err != nil {
		fmt.Println("Bad duration: ", err)
		return
	}
	r.wait = dur
	r.reconnect = true
}

//Connect establishes connection to specified address and stores encryption seed
//used by Login().
func (r *Rcon) Connect(address string) (err error) {
	r.service = address
	r.sock, err = net.Dial("tcp", address)
	if err != nil {
		return err
	}
	r.status = "connected"
	str := r.Scan("### Digest seed:")
	r.seed = strings.TrimSpace(strings.Split(str, ":")[1])
	return
}

//Login encrypts seed & pass, performs authentication with Rcon server.
func (r *Rcon) Login(admin, pass string) (err error) {
	r.pass = pass
	r.admin = admin
	hash := md5.New()
	hash.Write([]byte(r.seed + pass))
	_, err = r.sock.Write([]byte("login " + fmt.Sprintf("%x", hash.Sum(nil)) + "\n"))
	if err != nil {
		return err
	}
	r.Scan("Authentication successful")
	if len(r.admin) > 0 {
		r.Send(fmt.Sprintf("bf2cc setadminname %s", r.admin))
	}
	r.status = "authenticated"
	return
}

//Reconnect attempts to re-establish Rcon connection. Waiting duration & trying
//again on failure.
func (r *Rcon) Reconnect() error {
	for {
		r.status = "reconnecting"
		fmt.Println("Attempting reconnection.")
		if err := r.Connect(r.service); err != nil {
			fmt.Println("Reconnection attempt failed. Waiting.")
			time.Sleep(r.wait)
			continue
		}
		if err := r.Login(r.admin, r.pass); err != nil {
			return err
		}
		fmt.Println("Reconnection successful.")
		break
	}
	return nil
}

//Scan parses incoming socket data for specified string & returns the data found.
func (r *Rcon) Scan(str string) (s string) {
	scanner := bufio.NewScanner(r.sock)
	for scanner.Scan() {
		if s = scanner.Text(); strings.Contains(s, str) {
			fmt.Println(s)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return
}

//Send is a single-use style function, independant of the Reader & Writer, used
//to write a command to the socket and returning the resulting data as a string.
//Includes reconnection.
func (r *Rcon) Send(command string) (string, error) {
	line := "\u0002" + command + "\n"
	_, err := r.sock.Write([]byte(line))
	if err != nil {
		fmt.Println("Write/connection issue:", err)
		if strings.Contains(fmt.Sprintf("%s", err), "connect") && r.reconnect {
			if err := r.Reconnect(); err != nil {
				return "", err
			} else {
				return r.Send(command)
			}
		}
	}
	result, err := bufio.NewReader(r.sock).ReadString('\u0004')
	if err != nil {
		fmt.Println("Reader/connection issue:", err)
		if strings.Contains(fmt.Sprintf("%s", err), "connect") && r.reconnect {
			if err := r.Reconnect(); err != nil {
				return "", err
			} else {
				return r.Send(command)
			}
		}
	}
	return strings.TrimSpace(strings.Trim(result, "\u0004")), nil
}

//Reader reads all incoming socket data, handling reconnection if enabled.
func (r *Rcon) Reader() {
	for {
		result, err := bufio.NewReader(r.sock).ReadString('\u0004')
		if err != nil {
			fmt.Println(err)
			r.status = "error"
			if strings.Contains(fmt.Sprintf("%s", err), "connect") && r.reconnect {
				r.Reconnect()
			}
		}
		result = strings.TrimSpace(strings.Trim(result, "\u0004"))
		if len(result) > 0 {
			//fmt.Println(result)
			r.receive <- result
		}
	}
}

//Writer handles writing send channel data to the socket. Waits if connection is
//not authenticated yet.
func (r *Rcon) Writer() {
	r.send = make(chan []byte)
	for message := range r.send {
		for r.status != "authenticated" { //wait if not authenticated
			time.Sleep(1 * time.Second)
		}
		line := "\u0002" + fmt.Sprintf("%s", message) + "\n"
		_, err := r.sock.Write([]byte(line))
		if err != nil {
			fmt.Println(err)
			r.status = "error"
		}
	}
}

//Write sends a message to Rcon.send channel to be written out by Writer().
func (r *Rcon) Write(message string) {
	for r.send == nil { //wait to send if channel is not available
		time.Sleep(1 * time.Second)
	}
	r.send <- []byte(strings.TrimSpace(message))
}

//Init initializes Reader & Writer routines. Also initializes necessary channels
//and starts the Queue for handling outgoing commands with Enqueue().
func (r *Rcon) Init() {
	r.queue = make(chan string)
	r.receive = make(chan string)
	go r.Reader()
	go r.Writer()
	r.Queue()
}

//Enqueue adds a command line to the Queue to be written to the Rcon connection via Writer.
func (r *Rcon) Enqueue(line string) {
	for r.queue == nil { //wait to queue if channel is not available
		time.Sleep(1 * time.Second)
	}
	r.queue <- line
}

//Queue sequentially handles outgoing commands being sent to the Rcon connection.
func (r *Rcon) Queue() {
	for s := range r.queue {
		r.Write(s)
		time.Sleep(100 * time.Millisecond)
	}
}

//Handler listens on the receive channel for data from the Reader. Runs the given
//function on the resulting string data.
func (r *Rcon) Handler(f func(string)) {
	for s := range r.receive {
		f(s)
	}
}
