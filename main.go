/* gorcon version 13.12.23 (lee8oi)

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
	pass, seed, service string
	reconnect           bool
	sock                net.Conn
	send                chan []byte
}

//AutoReconnect is used to enable/disable automatic socket reconnection. True to enable.
func (r *Rcon) AutoReconnect(b bool) {
	r.reconnect = b
}

//Connect establishes connection to specified address and grabs encryption seed
//used by Login().
func (r *Rcon) Connect(address string) (err error) {
	r.service = address
	r.sock, err = net.Dial("tcp", address)
	if err != nil {
		return err
	}
	str := r.Scan("### Digest seed:")
	r.seed = strings.TrimSpace(strings.Split(str, ":")[1])
	return
}

//Login encrypts seed & pass, performs authentication.
func (r *Rcon) Login(pass string) (err error) {
	r.pass = pass
	hash := md5.New()
	hash.Write([]byte(r.seed + pass))
	_, err = r.sock.Write([]byte("login " + fmt.Sprintf("%x", hash.Sum(nil)) + "\n"))
	if err != nil {
		return err
	}
	r.Scan("Authentication successful")
	return
}

//Reader reads all incoming socket data, handling reconnection if enabled.
func (r *Rcon) Reader() {
	for {
		result, err := bufio.NewReader(r.sock).ReadString('\u0004')
		if err != nil {
			fmt.Println(err)
			if strings.Contains(fmt.Sprintf("%s", err), "connection") && r.reconnect {
				r.Reconnect(30 * time.Second)
			} else {
				break
			}
		}
		fmt.Println(strings.TrimSpace(strings.Trim(result, "\u0004")))
	}
}

//Reconnect attempts to re-establish Rcon connection. Waiting duration & trying
//again on failure.
func (r *Rcon) Reconnect(duration time.Duration) error {
	for {
		fmt.Println("Attempting reconnection.")
		if err := r.Connect(r.service); err != nil {
			fmt.Println("Reconnection attempt failed. Waiting.")
			time.Sleep(duration)
			continue
		}
		if err := r.Login(r.pass); err != nil {
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

//Send is a single-use style function for writing a command to the socket and
//returning the resulting data as a string.
func (r *Rcon) Send(command string) (string, error) {
	line := "\u0002" + command + "\n"
	_, err := r.sock.Write([]byte(line))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	result, err := bufio.NewReader(r.sock).ReadString('\u0004')
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return strings.TrimSpace(strings.Trim(result, "\u0004")), nil
}

//Writer handles writing send channel data to the socket, prefixing as needed to
//enable EOT to be appended to the resulting data.
func (r *Rcon) Writer() {
	r.send = make(chan []byte, 256)
	for message := range r.send {
		line := "\u0002" + fmt.Sprintf("%s", message) + "\n"
		_, err := r.sock.Write([]byte(line))
		if err != nil {
			fmt.Println(err)
		}
	}
}

//Write sends a message to Rcon.send channel to be written out by Writer().
func (r *Rcon) Write(message string) {
	r.send <- []byte(strings.TrimSpace(message))
}
