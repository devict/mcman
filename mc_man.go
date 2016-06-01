package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"gogs.bullercodeworks.com/brian/mc_man/util"
)

func main() {
	xmxVal := "1024M"
	xmsVal := "1024M"
	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "-help" {
			fmt.Println("Usage: mc_man <Xmx Value> <Xms Value>")
			fmt.Println("	<Xmx Value> - The Maximum Memory Allocation Pool for the JVM")
			fmt.Println("	<Xms Value> - The Initial Memory Allocation Pool")
			os.Exit(0)
		}
		if len(args) > 0 {
			xmxVal = args[0]
			if len(args) > 1 {
				xmsVal = args[1]
			}
		}
	}

	// The minecraft server command
	cmd := exec.Command("java", "-Xmx"+xmxVal, "-Xms"+xmsVal, "-jar", "minecraft_server.jar", "nogui")

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
		fmt.Println("mc_man stopped")
		util.StopServer = true
		close(ch)
	}()

	// Load the Config
	mm := util.NewManager(stdin)
	util.LoadConfig(&mm)

	go func() {
		for {
			s, ok := <-ch
			if !ok {
				break
			}
			m := util.NewMessage(s)
			fmt.Printf("\x1b[34;1m%s\x1b[0m", m.Output())
			mm.ProcessMessage(s)
		}
	}()

	// Web Server Routine
	go func() {
		util.StartServer(ch)
	}()

	// The forever loop to monitor everything
	for {
		time.Sleep(time.Second)
		if util.StopServer {
			break
		}
		//		fmt.Printf("Monitoring... (%d users)\n", len(util.GetConfig().LoggedInUsers))
		//		for i, u := range util.GetConfig().LoggedInUsers {
		//			if !u.HasQuota() {
		//				fmt.Printf(">> User %s is out of quota\n", u.Name)
		//			}
		//		}
	}
}
