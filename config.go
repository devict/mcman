package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

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
	MCUsers         []*MCUser
	LoggedInMCUsers []*MCUser
	Whitelist       []string
	Ops             []string
	TeleportPoints  map[string]string
	model           *Model
}

var c *Config

var StopServer = false
var mu sync.Mutex

var message_manager *MessageManager

func LoadConfig(mm *MessageManager) {
	message_manager = mm
	c = new(Config)

	c.model = InitializeModel()

	// Load the whitelist
	whitelist_rd, err := ioutil.ReadFile("whitelist.json")
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
	oplist_rd, err := ioutil.ReadFile("ops.json")
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

	config_str, err := ioutil.ReadFile("mcman.config")
	if err == nil {
		j, _ := jason.NewObjectFromBytes(config_str)
		o, _ := j.GetObjectArray("options")

		// Add the "Stop" listener
		fmt.Println("> Activating 'stop' listener")
		AddListener(func(i *Message) bool {
			if i.MCUser.IsOp && i.Text == "!stop\n" {
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
										SetPorch(i.MCUser.Name, p_str)
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
								porch_str, found := GetPorch(username)
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
				// The 'teleport to spawn' listener
				c.TeleportPoints = make(map[string]string)
				// Load all of the teleport points
				if point_array, err := option.GetObjectArray("points"); err == nil {
					if len(point_array) > 0 {
						for k := range point_array {
							if tp_name, err := point_array[k].GetString("name"); err == nil {
								if tp_loc, err := point_array[k].GetString("location"); err == nil {
									c.TeleportPoints[tp_name] = tp_loc
								}
							}
						}
					}
				}

				c.FeatureTP = opt_enabled
				if opt_enabled {
					fmt.Println("> Activating 'teleport' listener")
					// Add !tp listener
					AddListener(func(i *Message) bool {
						if i.MCUser.Name != "" && strings.HasPrefix(i.Text, "!tp ") {
							tp_name := strings.Split(i.Text, " ")[1]
							tp_str, found := GetTPPoint(tp_name)
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
										SetTPPoint(tp_name, p_str)
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
				if i.MCUser.Name == "" && strings.Contains(i.Text, " logged in with entity id ") {
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
							LoginMCUser(*FindMCUser(find, true))
							return true
						}
					}
				}
				return false
			})
			// Add logout listener
			AddListener(func(i *Message) bool {
				if i.MCUser.Name == "" && strings.Contains(i.Text, " lost connection: ") {
					// Find the user that just logged out
					r := strings.Split(i.Text, "]: ")
					find := ""
					if len(r) > 0 {
						find = r[1]
						r := strings.Split(find, " lost connection: ")
						if len(r) > 0 {
							find = r[0]
							// find should be the user name now
							LogoutMCUser(*FindMCUser(find, false))
							return true
						}
					}
				}
				return false
			})

			// Add !help listener
			AddListener(func(i *Message) bool {
				if i.MCUser.Name != "" && i.Text == "!help\n" {
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

		c.MCUsers = make([]*MCUser, 0, 10)
		u, _ := j.GetObjectArray("users")
		for _, user := range u {
			user_name, err := user.GetString("name")
			if err == nil && user_name != "" {
				user_home, _ := user.GetString("home")
				user_porch, _ := user.GetString("porch")
				us := NewMCUser(user_name)
				for _, un := range c.Ops {
					if un == user_name {
						us.IsOp = true
					}
				}
				us.Home = user_home
				us.Porch = user_porch
				c.model.updateMcUser(us)
				c.MCUsers = append(c.MCUsers, us)
			}
		}
		fmt.Printf("> Loaded %d Users\n", len(c.MCUsers))
	}
}

func DoStopServer() {
	mu.Lock()
	message_manager.Output("stop")
	WriteConfig()
	c.model.closeDatabase()
	StopServer = true
	mu.Unlock()
}

func LoginMCUser(u MCUser) {
	for _, user := range c.LoggedInMCUsers {
		if user.Name == u.Name {
			// User is already logged in
			return
		}
	}
	c.LoggedInMCUsers = append(c.LoggedInMCUsers, &u)
}

func LogoutMCUser(u MCUser) {
	for idx, user := range c.LoggedInMCUsers {
		if user.Name == u.Name {
			t := append(c.LoggedInMCUsers[:idx], c.LoggedInMCUsers[idx+1:]...)
			c.LoggedInMCUsers = make([]*MCUser, len(t))
			copy(c.LoggedInMCUsers, t)
			return
		}
	}
}

func AddMCUser(username string) {
	if username != "" {
		us := NewMCUser(username)
		fmt.Println("Adding new user: " + username)
		c.MCUsers = append(c.MCUsers, us)
		WriteConfig()
	}
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
		// Output all TP Points
		d = d + ",\"points\":["
		for k, v := range c.TeleportPoints {
			d = d + "{\"name\":\"" + k + "\",\"location:\"" + v + "\"}"
		}
		if len(c.TeleportPoints) > 0 {
			d = d[:len(d)-1]
		}
		d = d + "]"
	} else {
		d = d + "false"
	}
	d = d + "},{\"name\":\"daynight\",\"enabled\":"
	if c.FeatureDayNight {
		d = d + "true"
	} else {
		d = d + "false"
	}
	d = d + "}],\"users\":["
	// Output users array
	num_users := 0
	for _, u := range c.MCUsers {
		if num_users > 0 {
			d = d + ","
		}
		d = d + u.ToJSONString()
		num_users++
	}
	d = d + "]}"
	do := []byte(d)
	ioutil.WriteFile("mcman.config", do, 0664)
}

func SetHome(user, loc string) {
	u := FindMCUser(user, true)
	if u.Index != -1 {
		u.Home = strings.Replace(loc, "\n", "", -1)
		// Replace the user in the Users array
		c.MCUsers[u.Index] = u
		WriteConfig()
	}
}

func GetHome(user string) (string, bool) {
	u := FindMCUser(user, false)
	if u.Index == -1 || u.Home == "" {
		return "", false
	}
	return u.Home, true
}

func SetPorch(user, loc string) {
	u := FindMCUser(user, true)
	if u.Index != -1 {
		u.Porch = strings.Replace(loc, "\n", "", -1)
		c.MCUsers[u.Index] = u
		WriteConfig()
	}
}

func GetPorch(user string) (string, bool) {
	u := FindMCUser(user, false)
	if u.Index == -1 || u.Porch == "" {
		return "", false
	}
	return u.Porch, true
}

func SetTPPoint(tp_name, loc string) {
	c.TeleportPoints[tp_name] = loc
}

func GetTPPoint(tp_name string) (string, bool) {
	ret := c.TeleportPoints[tp_name]
	return ret, (ret != "")
}

func FindMCUser(name string, create bool) *MCUser {
	for _, user := range c.MCUsers {
		if user.Name == name {
			return user
		}
	}
	if create && name != "" {
		AddMCUser(name)
		return FindMCUser(name, false)
	}
	return NewMCUser("")
}

func GetConfig() *Config {
	return c
}
