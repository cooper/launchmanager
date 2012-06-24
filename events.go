package LaunchManager

import (
	"fmt"
	"process"
)

var eventHandlers = make(map[string]func(conn *connection, name string, params map[string]interface{}))

// assign handlers
func createEventHandlers() {
	eventHandlers["register"] = registerHandler
	eventHandlers["run"] = runHandler
}

// creates a process object for the connected process.
func registerHandler(conn *connection, name string, params map[string]interface{}) {
	pid := params["pid"].(float64)
	conn.process = process.CFromPID(int(pid))
}

// runs an executable file
// ["run", {"id": 0, "file":"/bin/bash",argv:["some","arguments"]}]
func runHandler(conn *connection, name string, params map[string]interface{}) {

	// ensure that all values are present and of the right type.
	var args = map[string]string{
		"file": "string",
		"id":   "float64",
		"argv": "[]interface {}",
	}
	for arg, typ := range args {
		if params[arg] == nil || fmt.Sprintf("%T", params[arg]) != typ {
			return
		}
	}

	// extract interface values
	file := params["file"].(string)
	argv := params["argv"].([]interface{})
	id := params["id"].(float64)

	// run as root?
	asroot := false
	if params["asroot"].(bool) == true {
		asroot = true
	}

	// convert argv
	newargv := make([]string, len(argv)+1)
	newargv[0] = file
	for index, val := range argv {
		newargv[index+1] = val.(string)
	}

	// launch the process in a new goroutine
	go launchProcess(conn, int(id), file, newargv, asroot)
}
