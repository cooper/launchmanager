package LaunchManager

import (
	"bufio"
	"encoding/json"
	"net"
	"process"
)

var currentId int = 0

type connection struct {
	socket   *net.UnixConn
	incoming *bufio.Reader
	id       int
	process  *process.CProcess
}

// create a new connection
func newConnection(conn *net.UnixConn) *connection {
	currentId++
	newconn := &connection{
		socket:   conn,
		incoming: bufio.NewReader(conn),
		id:       currentId,
	}
	return newconn
}

// read data from a connection
func (conn *connection) readData() {
	for {
		line, _, err := conn.incoming.ReadLine()
		if err != nil {
			return
		}
		handleEvent(conn, line)
	}
}

// handle a JSON event
func handleEvent(conn *connection, data []byte) bool {
	var i interface{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return false
	}

	// should be an array.
	c := i.([]interface{}) // type assertion -- because an interface is a container

	command := c[0].(string)
	params := c[1].(map[string]interface{})

	// if a handler for this command exists, run it
	if eventHandlers[command] != nil {
		eventHandlers[command](conn, command, params)
	}

	return true
}

// send a JSON event
func (conn *connection) send(command string, params map[string]interface{}) bool {
	b, err := json.Marshal(params)
	if err != nil {
		return false
	}
	b = append(b, '\n')
	_, err = conn.socket.Write(b)
	if err != nil {
		return false
	}
	return true
}
