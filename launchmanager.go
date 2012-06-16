package LaunchManager

import (
	"errors"
	"net"
	"os"
	"os/exec"
)

func Run() (err error) {
	const path = "/system/socket/LaunchSocket"

	if os.Getpid() != 1 {
		return errors.New("invalid process ID")
	}

	// launch sysinit and wait for it to exit
	launchFirst("/system/executable/sysinit")

	// check if file exists. if so, delete it.
	if _, err := os.Lstat(path); err == nil {
		os.Remove(path)
	}

	// resolve the address
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return err
	}

	// listen on path
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return err
	}

	// create event handlers
	createEventHandlers()

	// run post-init programs
	launchFirst("/system/executable/postinit")

	// loop for connections
	for {

		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go newConnection(conn).readData()

	}
	return
}

// launch a process and wait for it to exit, replying to the connection
// with its status
func launchProcess(conn *connection, id int, file string, argv []string) {
	var err error

	// run the command
	cmd := exec.Command(file)
	cmd.Args = argv
	err = cmd.Start()

	// immediate error
	if err != nil {
		conn.send("processEnded", map[string]interface{}{
			"id":    id,
			"pid":   cmd.Process.Pid,
			"error": true,
		})
		return
	}

	// wait for the process to exit
	cmd.Wait()

	// send a successful "process ran" response
	conn.send("processEnded", map[string]interface{}{
		"id":    id,
		"pid":   cmd.Process.Pid,
		"error": false,
	})
}

// launch initial processes
func launchFirst(proc string) {
	cmd := exec.Command(proc)
	cmd.Run()
}
