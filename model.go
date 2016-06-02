package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

type Model struct {
	// Bucket Names
	mc_bucket        string
	mc_users_bucket  string
	mc_config_bucket string
	web_bucket       string
	web_users_bucket string

	// Key prefixes
	user_prefix           string
	config_feature_prefix string

	db *bolt.DB
}

func InitializeModel() *Model {
	var err error
	ret := new(Model)
	ret.mc_bucket = "mc"
	ret.mc_users_bucket = "mc_users"
	ret.mc_config_bucket = "mc_config"
	ret.web_bucket = "web"
	ret.web_users_bucket = "web_users"
	ret.user_prefix = "user_"
	ret.config_feature_prefix = "feature_"

	ret.db, err = bolt.Open("mcman.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func (m *Model) closeDatabase() {
	m.db.Close()
}

/* Web Server Stuff */
func (m *Model) getWebUser(username string) WebUser {
	var ret WebUser
	m.db.View(func(tx *bolt.Tx) error {
		if web_b := tx.Bucket([]byte(m.web_bucket)); web_b != nil {
			if web_u_b := web_b.Bucket([]byte(m.web_users_bucket)); web_u_b != nil {
				user_key := m.user_prefix + username
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

func (m *Model) updateWebUser(u *WebUser) {
	fmt.Printf("BOLT: Adding Web User %s\n", u.Username)
	m.db.Update(func(tx *bolt.Tx) error {
		web_b, err := tx.CreateBucketIfNotExists([]byte(m.web_bucket))
		if err != nil {
			return err
		}

		web_u_b, err := web_b.CreateBucketIfNotExists([]byte(m.web_users_bucket))
		if err != nil {
			return err
		}
		user_key := m.user_prefix + u.Username
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
	err := m.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(m.mc_bucket))
		if err != nil {
			return err
		}
		bc, err := b.CreateBucketIfNotExists([]byte(m.mc_config_bucket))
		if err != nil {
			return err
		}
		addBooleanPairToBucket(bc, m.config_feature_prefix+opt, enabled)
		return nil
	})
	if err != nil {
		fmt.Printf("Save Feature Error: %s", err)
	}
}

func (m *Model) mcFeatureIsEnabled(opt string) bool {
	ret := false
	m.db.View(func(tx *bolt.Tx) error {
		lookingfor := []byte(opt)
		b := tx.Bucket([]byte(m.mc_bucket))
		if b != nil {
			bc := b.Bucket([]byte(m.mc_config_bucket))
			c := bc.Cursor()
			srch_prefix := []byte(m.config_feature_prefix + opt)
			fmt.Printf("%s:%s=> ", string(m.mc_config_bucket), string(srch_prefix))
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
	var ret []MCUser
	m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(m.mc_users_bucket))
		c := b.Cursor()
		srch_prefix := []byte(m.user_prefix)
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
	fmt.Printf("BOLT: Adding User %s\n", u.Name)
	m.db.Update(func(tx *bolt.Tx) error {
		mc_b, err := tx.CreateBucketIfNotExists([]byte(m.mc_bucket))
		if err != nil {
			return err
		}

		mc_u_b, err := mc_b.CreateBucketIfNotExists([]byte(m.mc_users_bucket))
		if err != nil {
			return err
		}
		user_key := m.user_prefix + u.Name
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
