package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var output = make(chan string)

func main() {
	xmx := flag.String("xmx", "1024M", "The Maximum Memory Allocation Pool for the JVM")
	xms := flag.String("xms", "1024M", "The Initial Memory Allocation Pool")
	dir := flag.String("dir", ".", "Path to the MC server directory")
	jar := flag.String("jar", "minecraft_server.jar", "Path to the MC server JAR relative from -dir")

	flag.Parse()

	// The minecraft server command
	cmd := exec.Command("java", "-Xmx"+*xmx, "-Xms"+*xms, "-d64", "-jar", *jar, "nogui")
	cmd.Dir = *dir

	// Control StdIn
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Pipe the process' stdout to stdout for monitoring
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Kick off the process
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// A routine to keep the channel updated
	go func() {
		var hasStarted bool
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n != 0 {
				// Something there throw it into the channel
				output <- string(buf[:n])
			}
			if err != nil {
				fmt.Print("Caught: ", err, "\n")
				// Error, break out of loop
				break
			}
			hasStarted = true
		}
		fmt.Println("mcman stopping")
		if !hasStarted {
			fmt.Println("Couldn't get things running. Is minecraft available?")
		}
		DoStopServer()
		close(output)
	}()

	// Load the Config
	mm := NewManager(stdin)
	LoadConfig(&mm, *dir)

	go func() {
		for s := range output {
			m := NewMessage(s)
			fmt.Printf("\x1b[34;1m%s\x1b[0m", m.Output())
			mm.ProcessMessage(m)
		}
	}()

	// Web Server Routine
	go func() {
		StartServer(true)
	}()

	// Catch interrupt signals and gracefully stop the servers
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println("Caught signal", sig)
		DoStopServer()
	}()

	// The forever loop to monitor everything
	for {
		time.Sleep(time.Second)

		mu.Lock()
		if StopServer {
			mu.Unlock()
			break
		}
		mu.Unlock()
	}
}
