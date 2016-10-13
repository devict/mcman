package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/antonholmquist/jason"
)

type Config struct {
	// The JSON object of what was read
	LoadedJson      jason.Object
	Options         jason.Object
	FeatureTPHome   bool
	FeatureTPVisit  bool
	FeatureTP       bool
	FeatureDayNight bool
	LoggedInMCUsers []MCUser
	Whitelist       []string
	Ops             []string
	dir             string
	model           *Model
}

var c *Config

var StopServer = false
var mu sync.Mutex

var message_manager *MessageManager

func LoadConfig(mm *MessageManager, dir string) {
	message_manager = mm
	c = new(Config)
	c.dir = dir
	c.model = InitializeModel()
	// If we don't have any web users yet, we need to create one
	allUsers := c.model.getAllWebUsers()
	if len(allUsers) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Create new Web User")
		fmt.Print("Username: ")
		uName, _ := reader.ReadString('\n')
		uName = strings.TrimSpace(uName)
		var pw1, pw2 []byte
		for string(pw1) != string(pw2) || string(pw1) == "" {
			fmt.Print("Password: ")
			pw1, _ = terminal.ReadPassword(0)
			fmt.Println("")
			fmt.Print("Repeat Password: ")
			pw2, _ = terminal.ReadPassword(0)
			fmt.Println("")
			if string(pw1) != string(pw2) {
				fmt.Println("Entered Passwords didn't match!")
			}
		}
		if err := c.model.updateWebUser(&WebUser{Username: uName, Password: string(pw1)}); err != nil {
			log.Fatal(err)
		}
	}

	// Load the whitelist
	whitelist_rd, err := ioutil.ReadFile(c.dir + "/whitelist.json")
	// We have to make it an object to read it...
	whitelist_rd = append(append([]byte("{\"whitelist\":"), whitelist_rd...), '}')
	if err == nil {
		j, _ := jason.NewObjectFromBytes(whitelist_rd)
		jo, _ := j.GetObjectArray("whitelist")
		for _, wl_u := range jo {
			n, _ := wl_u.GetString("name")
			fmt.Print("> Whitelisted User ", n, "\n")
			c.Whitelist = append(c.Whitelist, n)
		}
	}

	// Load the Op list
	oplist_rd, err := ioutil.ReadFile(c.dir + "/ops.json")
	// We have to make it an object to read it...
	oplist_rd = append(append([]byte("{\"ops\":"), oplist_rd...), '}')
	if err == nil {
		j, _ := jason.NewObjectFromBytes(oplist_rd)
		jo, _ := j.GetObjectArray("ops")
		for _, ol_u := range jo {
			n, _ := ol_u.GetString("name")
			fmt.Print("> Opped User ", n, "\n")
			c.Ops = append(c.Ops, n)
		}
	}

	config_str, err := ioutil.ReadFile(c.dir + "/mcman.config")
	if err == nil {
		j, _ := jason.NewObjectFromBytes(config_str)
		o, _ := j.GetObjectArray("options")

		// Add the "Stop" listener
		fmt.Println("> Activating 'stop' listener")
		AddListener(func(i *Message) bool {
			if i.MCUser != nil && i.Text == "!stop\n" {
				DoStopServer()
				return true
			}
			return false
		})

		c.FeatureTPHome = c.model.mcFeatureIsEnabled("tphome")
		if c.FeatureTPHome {
			fmt.Println("> Activating 'home' listeners")
			// Add !set home listener
			AddListener(func(i *Message) bool {
				if i.MCUser != nil && i.Text == "!set home\n" {
					AddTempListener(func(inp *Message) bool {
						listen_for := "Teleported " + i.MCUser.Name + " to "
						if inp.MCUser.Name == "" && strings.Contains(inp.Text, listen_for) {
							// Found the text
							r := strings.Split(inp.Text, listen_for)
							if len(r) > 0 {
								p_str := r[1]
								p_str = strings.Replace(p_str, ",", "", -1)
								p_str = strings.Replace(p_str, "\n", "", -1)
								/* TODO: Update User's Home in DB */
								mm.Tell(i.MCUser.Name, "Set your home to "+p_str, "blue")
								return true
							}
						}
						return false
					})
					mm.Output("tp " + i.MCUser.Name + " ~ ~ ~")
					return true
				}
				return false
			})
			// Add !home listener
			AddListener(func(i *Message) bool {
				if i.MCUser != nil && i.Text == "!home\n" {
					// TODO: Lookup home in DB
					home_str, found := "", false
					if found {
						mm.Output("tp " + i.MCUser.Name + " " + home_str)
					} else {
						mm.Tell(i.MCUser.Name, "I don't know where your home is. Set it to your current position by typing '!set home'", "red")
					}
				}
				return false
			})
		}

		for _, option := range o {
			opt_name, _ := option.GetString("name")
			opt_enabled, _ := option.GetBoolean("enabled")
			if opt_name == "home" {
				/*
					c.FeatureTPHome = opt_enabled
					if opt_enabled {
						fmt.Println("> Activating 'home' listeners")
						// Add !set home listener
						AddListener(func(i *Message) bool {
							if i.MCUser.Name != "" && i.Text == "!set home\n" {
								AddTempListener(func(inp *Message) bool {
									listen_for := "Teleported " + i.MCUser.Name + " to "
									if inp.MCUser.Name == "" && strings.Contains(inp.Text, listen_for) {
										// Found the text
										r := strings.Split(inp.Text, listen_for)
										if len(r) > 0 {
											p_str := r[1]
											p_str = strings.Replace(p_str, ",", "", -1)
											p_str = strings.Replace(p_str, "\n", "", -1)
											SetHome(i.MCUser.Name, p_str)
											mm.Tell(i.MCUser.Name, "Set your home to "+p_str, "blue")
											return true
										}
									}
									return false
								})
								mm.Output("tp " + i.MCUser.Name + " ~ ~ ~")
								return true
							}
							return false
						})
						// Add !home listener
						AddListener(func(i *Message) bool {
							if i.MCUser.Name != "" && i.Text == "!home\n" {
								home_str, found := GetHome(i.MCUser.Name)
								if found {
									mm.Output("tp " + i.MCUser.Name + " " + home_str)
								} else {
									mm.Tell(i.MCUser.Name, "I don't know where your home is. Set it to your current position by typing '!set home'", "red")
								}
							}
							return false
						})
					}
				*/
			} else if opt_name == "visit" {
				c.FeatureTPVisit = opt_enabled
				if opt_enabled {
					fmt.Println("> Activating 'visit' listeners")
					// Add !set porch listener
					AddListener(func(i *Message) bool {
						if i.MCUser.Name != "" && i.Text == "!set porch\n" {
							AddTempListener(func(inp *Message) bool {
								listen_for := "Teleported " + i.MCUser.Name + " to "
								if inp.MCUser.Name == "" && strings.Contains(inp.Text, listen_for) {
									// Found the text
									r := strings.Split(inp.Text, listen_for)
									if len(r) > 0 {
										p_str := r[1]
										p_str = strings.Replace(p_str, ",", "", -1)
										// TODO: Set Porch in DB
										mm.Tell(i.MCUser.Name, "Set your porch to "+p_str, "blue")
										return true
									}
								}
								return false
							})
							mm.Output("tp " + i.MCUser.Name + " ~ ~ ~")
							return true
						}
						return false
					})
					// Add !visit listener
					AddListener(func(i *Message) bool {
						if i.MCUser.Name != "" && strings.HasPrefix(i.Text, "!visit ") {
							// Find the user we're trying to visit
							r := strings.Split(strings.Replace(i.Text, "\n", "", -1), "!visit ")
							if len(r) > 0 {
								username := r[1]
								// TODO: Get Porch from DB
								porch_str, found := "", false
								if found {
									mm.Output("tp " + i.MCUser.Name + " " + porch_str)
								} else {
									mm.Tell(i.MCUser.Name, "I don't know where "+username+"'s porch is. They can set it to their current position by typing '!set porch'", "red")
								}
							}
						}
						return false
					})
				}
			} else if opt_name == "teleport" {
				c.FeatureTP = opt_enabled
				if opt_enabled {
					fmt.Println("> Activating 'teleport' listener")
					// Add !tp listener
					AddListener(func(i *Message) bool {
						if i.MCUser.Name != "" && strings.HasPrefix(i.Text, "!tp ") {
							tp_name := strings.Split(i.Text, " ")[1]
							// TODO: Get Teleport Point from DB
							tp_str, found := "", false
							if found {
								mm.Output("tp " + i.MCUser.Name + " " + tp_str)
							} else {
								mm.Tell(i.MCUser.Name, "I don't know where "+tp_name+".", "red")
							}
						}
						return false
					})
					// Add !tpset listener
					AddListener(func(i *Message) bool {
						if i.MCUser.IsOp && strings.HasPrefix(i.Text, "!tpset ") {
							tp_name := strings.Split(i.Text, " ")[1]
							// Save user's current position as tp_name point
							AddTempListener(func(inp *Message) bool {
								listen_for := "Teleported " + i.MCUser.Name + " to "
								if inp.MCUser.Name == "" && strings.Contains(inp.Text, listen_for) {
									// Found the text
									r := strings.Split(inp.Text, listen_for)
									if len(r) > 0 {
										p_str := r[1]
										p_str = strings.Replace(p_str, ",", "", -1)
										// TODO: Set the Teleport Point in the DB
										mm.Tell(i.MCUser.Name, "Added TP Point "+tp_name+" at "+p_str, "blue")
										return true
									}
								}
								return false
							})
							mm.Output("tp " + i.MCUser.Name + " ~ ~ ~")
							return true
						}
						return false
					})
				}
			} else if opt_name == "daynight" {
				c.FeatureDayNight = opt_enabled
				if opt_enabled {
					fmt.Println("> Activating 'time' listeners")
					// Add !time day listener
					AddListener(func(i *Message) bool {
						if i.MCUser.Name != "" && i.Text == "!time day\n" {
							// TODO: Start vote
							mm.Output("time set day")
							mm.Tell("@a", "Day Time time initiated by "+i.MCUser.Name, "yellow")
							return true
						}
						return false
					})
					// Add !time night listener
					AddListener(func(i *Message) bool {
						if i.MCUser.Name != "" && i.Text == "!time night\n" {
							// TODO: Start vote
							mm.Output("time set night")
							mm.Tell("@a", "Night Time time initiated by "+i.MCUser.Name, "blue")
							return true
						}
						return false
					})
				}
			}
			// Add login listener
			AddListener(func(i *Message) bool {
				if i.MCUser == nil && strings.Contains(i.Text, " logged in with entity id ") {
					// TODO: User Logged in Function
					// Find the user that just logged in
					r := strings.Split(i.Text, "]: ")
					find := ""
					if len(r) > 0 {
						find = r[1]
						r := strings.Split(find, "[/")
						if len(r) > 0 {
							find = r[0]
							// find should be the user name now
							var u *MCUser
							var err error
							if u, err = c.model.getMCUser(find); err != nil {
								// user doesn't exist, create it
								u = new(MCUser)
								u.Name = find
							}
							u.LoginTime = time.Now()
							c.model.updateMCUser(u)
							return true
						}
					}
				}
				return false
			})
			// Add logout listener
			AddListener(func(i *Message) bool {
				if i.MCUser == nil && strings.Contains(i.Text, " lost connection: ") {
					// Find the user that just logged out
					r := strings.Split(i.Text, "]: ")
					find := ""
					if len(r) > 0 {
						find = r[1]
						r := strings.Split(find, " lost connection: ")
						if len(r) > 0 {
							find = r[0]
							// find should be the user name now
							var u *MCUser
							var err error
							if u, err = c.model.getMCUser(find); err != nil {
								// user doesn't exist, create it
								u = new(MCUser)
								u.Name = find
							}
							u.LogoutTime = time.Now()
							c.model.updateMCUser(u)
							return true
						}
					}
				}
				return false
			})

			// Add !help listener
			AddListener(func(i *Message) bool {
				if i.MCUser != nil && i.Text == "!help\n" {
					mm.Tell(i.MCUser.Name, "-=( mcman Manager Help )=-", "blue")
					numFeatures := 0
					if c.FeatureTPHome == true {
						numFeatures++
						mm.Tell(i.MCUser.Name, "!set home -- Set your 'home' to your current position.", "white")
						mm.Tell(i.MCUser.Name, "!home -- Request a teleport to your 'home' position.", "white")
					}
					if c.FeatureTPVisit == true {
						numFeatures++
						mm.Tell(i.MCUser.Name, "!set porch -- Set your 'porch' to your current position.", "white")
						mm.Tell(i.MCUser.Name, "!visit <username> -- Request a teleport to <username>'s 'porch' position.", "white")
					}
					if c.FeatureDayNight == true {
						numFeatures++
						mm.Tell(i.MCUser.Name, "!time day -- Ask the server to time the time to 'day'.", "white")
						mm.Tell(i.MCUser.Name, "!time night -- Ask the server to time the time to 'night'.", "white")
					}
					if numFeatures == 0 {
						mm.Tell(i.MCUser.Name, "mcman currently has no user features loaded.", "white")
					}
					mm.Tell(i.MCUser.Name, "-=========================-", "blue")
					return true
				}
				return false
			})
		}

		if allUsers, err := c.model.getAllMCUsers(); err != nil {
			fmt.Printf("> Error loading users: " + err.Error())
		} else {
			fmt.Printf("> Loaded %d Users\n", len(allUsers))
		}
	}
}

func DoStopServer() {
	mu.Lock()
	message_manager.Output("stop")
	// Mark all Online users as Logged Out
	if ou, err := c.model.getOnlineMCUsers(); err == nil {
		for i := range ou {
			ou[i].LogoutTime = time.Now()
			c.model.updateMCUser(&ou[i])
		}
	}
	WriteConfig()
	StopServer = true
	mu.Unlock()
}

func WriteConfig() {
	c.model.mcSaveFeature("tphome", c.FeatureTPHome)
	c.model.mcSaveFeature("tpvisit", c.FeatureTPVisit)
	c.model.mcSaveFeature("tp", c.FeatureTP)
	c.model.mcSaveFeature("daynight", c.FeatureDayNight)

	// TODO: Make mcman aware of the world
	// Generate the JSON string for the config file
	d := "{\"options\":["
	// Output options array
	d = d + "{\"name\":\"home\",\"enabled\":"
	if c.FeatureTPHome {
		d = d + "true"
	} else {
		d = d + "false"
	}
	d = d + "},{\"name\":\"visit\",\"enabled\":"
	if c.FeatureTPVisit {
		d = d + "true"
	} else {
		d = d + "false"
	}
	d = d + "},{\"name\":\"teleport\",\"enabled\":"
	if c.FeatureTP {
		d = d + "true"
	} else {
		d = d + "false"
	}
	d = d + "},{\"name\":\"daynight\",\"enabled\":"
	if c.FeatureDayNight {
		d = d + "true"
	} else {
		d = d + "false"
	}
	d = d + "}]}"
	do := []byte(d)
	ioutil.WriteFile(c.dir+"/mcman.config", do, 0664)
}
