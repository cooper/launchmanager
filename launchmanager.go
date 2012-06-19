package LaunchManager

import (
	"errors"
	"libclient"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func Run() (err error) {
	const path = "/system/socket/LaunchSocket"

	// must run as root
	if os.Getuid() != 0 {
		return errors.New("must be run as root")
	}

	// must run as init
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

	// set permission to 777
	os.Chmod(path, 0777)

	// create event handlers
	createEventHandlers()

	// start and connect to ProcessManager
	go connectProcessManager()

	// run post-init programs
	launchFirst("/system/executable/postinit")

	// loop for connections
	for {

		conn, err := listener.AcceptUnix()
		if err != nil {
			return err
		}

		go newConnection(conn).readData()

	}
	return
}

// start and connect process manager
func connectProcessManager() {
	for {
		go launchFirst("/system/executable/ProcessManager")
		time.Sleep(5)
		libclient.RunProcess(map[string]string{
			"name":    "LaunchManager",
			"version": "1.0",
		})
		time.Sleep(5)
	}
}

// launch a process and wait for it to exit, replying to the connection
// with its status
func launchProcess(conn *connection, id int, file string, argv []string) {
	var err error

	// run the command
	cmd := exec.Command(file)
	cmd.Args = argv
	syscall.Setuid(1000)
	err = cmd.Start()
	syscall.Setuid(0)

	// immediate error
	if err != nil {
		conn.send("processEnded", map[string]interface{}{
			"id":    id,
			"pid":   cmd.Process.Pid,
			"error": true,
		})
		fmt.Println(err.Error())
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
