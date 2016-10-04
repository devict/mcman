package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/boltdb/bolt"
)

type Model struct {
	// Bucket Names
	mcBucket       string
	mcUsersBucket  string
	mcConfigBucket string
	webBucket      string
	webUsersBucket string

	// Key prefixes
	userPrefix          string
	configFeaturePrefix string

	db *bolt.DB
}

func InitializeModel() *Model {
	ret := new(Model)
	ret.mcBucket = "mc"
	ret.mcUsersBucket = "mc_users"
	ret.mcConfigBucket = "mc_config"
	ret.webBucket = "web"
	ret.webUsersBucket = "web_users"
	ret.userPrefix = "user_"
	ret.configFeaturePrefix = "feature_"

	// Make sure we can access the DB
	ret.openDatabase()
	ret.closeDatabase()
	allUsers := ret.getAllWebUsers()
	for i := range allUsers {
		fmt.Println(">> " + allUsers[i].Username)
	}
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
		if err := ret.updateWebUser(&WebUser{Username: uName, Password: string(pw1)}); err != nil {
			log.Fatal(err)
		}
	}

	return ret
}

func (m *Model) openDatabase() {
	var err error
	m.db, err = bolt.Open("mcman.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (m *Model) closeDatabase() {
	m.db.Close()
}

/* Web Server Stuff */
func (m *Model) getAllWebUsers() []WebUser {
	m.openDatabase()
	defer m.closeDatabase()

	var ret []WebUser
	m.db.View(func(tx *bolt.Tx) error {
		if webB := tx.Bucket([]byte(m.webBucket)); webB != nil {
			c := webB.Cursor()
			srchPrefix := []byte(m.userPrefix)
			for k, _ := c.Seek(srchPrefix); bytes.HasPrefix(k, srchPrefix); k, _ = c.Next() {
				if webUB := webB.Bucket(k); webUB != nil {
					retUName := string(webUB.Get([]byte("username")))
					retPass := string(webUB.Get([]byte("password")))
					ret = append(ret, WebUser{Username: retUName, Password: retPass})
				}
			}
		}
		return errors.New("No Web Users")
	})
	return ret
}

func (m *Model) getWebUser(username string) WebUser {
	m.openDatabase()
	defer m.closeDatabase()

	var ret WebUser
	m.db.View(func(tx *bolt.Tx) error {
		if web_b := tx.Bucket([]byte(m.webBucket)); web_b != nil {
			if web_u_b := web_b.Bucket([]byte(m.webUsersBucket)); web_u_b != nil {
				user_key := m.userPrefix + username
				if ub := web_u_b.Bucket([]byte(user_key)); ub != nil {
					ret_uname := string(ub.Get([]byte("username")))
					ret_pass := string(ub.Get([]byte("password")))
					ret = WebUser{Username: ret_uname, Password: ret_pass}
					return nil
				} else {
					return errors.New("Invalid User")
				}
			}
		}
		return errors.New("No Web Users")
	})
	return ret
}

func (m *Model) updateWebUser(u *WebUser) error {
	m.openDatabase()
	defer m.closeDatabase()

	fmt.Printf("BOLT: Adding/updating Web User %s\n", u.Username)
	return m.db.Update(func(tx *bolt.Tx) error {
		web_b, err := tx.CreateBucketIfNotExists([]byte(m.webBucket))
		if err != nil {
			return err
		}

		web_u_b, err := web_b.CreateBucketIfNotExists([]byte(m.webUsersBucket))
		if err != nil {
			return err
		}
		user_key := m.userPrefix + u.Username
		ub, uberr := web_u_b.CreateBucketIfNotExists([]byte(user_key))
		if uberr != nil {
			return uberr
		}
		addStringPairToBucket(ub, "username", u.Username)
		addStringPairToBucket(ub, "password", u.Password)
		return nil
	})
}

func (m *Model) mcSaveFeature(opt string, enabled bool) {
	m.openDatabase()
	defer m.closeDatabase()

	err := m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(m.mcBucket))
		if err != nil {
			return err
		}
		bc, err := b.CreateBucketIfNotExists([]byte(m.mcConfigBucket))
		if err != nil {
			return err
		}
		addBooleanPairToBucket(bc, m.configFeaturePrefix+opt, enabled)
		return nil
	})
	if err != nil {
		fmt.Printf("Save Feature Error: %s", err)
	}
}

func (m *Model) mcFeatureIsEnabled(opt string) bool {
	m.openDatabase()
	defer m.closeDatabase()

	ret := false
	m.db.View(func(tx *bolt.Tx) error {
		lookingfor := []byte(opt)
		b := tx.Bucket([]byte(m.mcBucket))
		if b != nil {
			bc := b.Bucket([]byte(m.mcConfigBucket))
			c := bc.Cursor()
			srch_prefix := []byte(m.configFeaturePrefix + opt)
			fmt.Printf("%s:%s=> ", string(m.mcConfigBucket), string(srch_prefix))
			for k, v := c.Seek(srch_prefix); bytes.HasPrefix(k, srch_prefix); k, v = c.Next() {
				// k should be the feature name, v is whether it is enabled or not
				fmt.Printf("%s == %s => ", string(k), string(lookingfor))
				if bytes.Equal(k, lookingfor) {
					ret = bytes.Equal(v, []byte("true"))
					if ret {
						fmt.Printf("It's On!\n")
					} else {
						fmt.Printf("It's Off!\n")
					}
					return nil
				}
			}
		}
		return errors.New("Feature Not Found")
	})
	return ret
}

/* Minecraft Config Stuff */
func (m *Model) getMcUsers() []MCUser {
	m.openDatabase()
	defer m.closeDatabase()

	var ret []MCUser
	m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(m.mcUsersBucket))
		c := b.Cursor()
		srch_prefix := []byte(m.userPrefix)
		for k, _ := c.Seek(srch_prefix); bytes.HasPrefix(k, srch_prefix); k, _ = c.Next() {
			if user_bucket := b.Bucket(k); user_bucket != nil {
				if us_name := user_bucket.Get([]byte("name")); us_name != nil {
					new_user := NewMCUser(string(us_name))
					new_user.IsOp = bytes.Equal(user_bucket.Get([]byte("op")), []byte("true"))
					new_user.Home = string(user_bucket.Get([]byte("home")))
					new_user.Porch = string(user_bucket.Get([]byte("porch")))
					new_user.Quota, _ = time.ParseDuration(string(user_bucket.Get([]byte("quota"))))
					new_user.quotaUsed, _ = time.ParseDuration(string(user_bucket.Get([]byte("quota_used"))))
					ret = append(ret, *new_user)
				}
			}
		}
		return nil
	})
	return ret
}

// updateMcUser adds or updates a user
func (m *Model) updateMcUser(u *MCUser) {
	m.openDatabase()
	defer m.closeDatabase()

	fmt.Printf("BOLT: Adding User %s\n", u.Name)
	m.db.Update(func(tx *bolt.Tx) error {
		mc_b, err := tx.CreateBucketIfNotExists([]byte(m.mcBucket))
		if err != nil {
			return err
		}

		mc_u_b, err := mc_b.CreateBucketIfNotExists([]byte(m.mcUsersBucket))
		if err != nil {
			return err
		}
		user_key := m.userPrefix + u.Name
		ub, uberr := mc_u_b.CreateBucketIfNotExists([]byte(user_key))
		if uberr != nil {
			return uberr
		}
		addStringPairToBucket(ub, "name", u.Name)
		addBooleanPairToBucket(ub, "op", u.IsOp)
		addStringPairToBucket(ub, "home", u.Home)
		addStringPairToBucket(ub, "porch", u.Porch)
		addDurationPairToBucket(ub, "quota", u.Quota)
		addDurationPairToBucket(ub, "quotaused", u.quotaUsed)
		addTimePairToBucket(ub, "logintime", u.loginTime)
		return nil
	})
}

func addStringPairToBucket(b *bolt.Bucket, k, v string) error {
	if err := b.Put([]byte(k), []byte(v)); err != nil {
		return err
	}
	return nil
}

func addBooleanPairToBucket(b *bolt.Bucket, k string, v bool) error {
	write_v := "true"
	if !v {
		write_v = "false"
	}
	if err := b.Put([]byte(k), []byte(write_v)); err != nil {
		return err
	}
	return nil
}

func addDurationPairToBucket(b *bolt.Bucket, k string, v time.Duration) error {
	write_v := v.String()
	if err := b.Put([]byte(k), []byte(write_v)); err != nil {
		return err
	}
	return nil
}

func addTimePairToBucket(b *bolt.Bucket, k string, v time.Time) error {
	if err := b.Put([]byte(k), []byte(v.String())); err != nil {
		return err
	}
	return nil
}
