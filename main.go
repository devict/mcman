package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"
)

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

	// Build a string channel for stdout
	ch := make(chan string)

	// And a routine to keep the channel updated
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n != 0 {
				// Something there throw it into the channel
				ch <- string(buf[:n])
			}
			if err != nil {
				fmt.Print("Caught: ", err, "\n")
				// Error, break out of loop
				break
			}
		}
		fmt.Println("mcman stopping")

		DoStopServer()
		close(ch)
	}()

	// Load the Config
	mm := NewManager(stdin)
	LoadConfig(&mm)

	go func() {
		for {
			s, ok := <-ch
			if !ok {
				break
			}
			m := NewMessage(s)
			fmt.Printf("\x1b[34;1m%s\x1b[0m", m.Output())
			mm.ProcessMessage(s)
		}
	}()

	// Web Server Routine
	go func() {
		StartServer(ch)
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
		//		fmt.Printf("Monitoring... (%d users)\n", len(GetConfig().LoggedInUsers))
		//		for i, u := range GetConfig().LoggedInUsers {
		//			if !u.HasQuota() {
		//				fmt.Printf(">> User %s is out of quota\n", u.Name)
		//			}
		//		}
	}
}
