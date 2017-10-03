package main

import (
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/br0xen/boltease"
)

type Model struct {
	// Bucket Paths
	mcBucket       []string
	mcUsersBucket  []string
	mcConfigBucket []string
	webBucket      []string
	webUsersBucket []string

	// Key prefixes
	userPrefix          string
	configFeaturePrefix string

	db *boltease.DB
}

func InitializeModel() *Model {
	ret := new(Model)
	ret.mcBucket = []string{"mc"}
	ret.mcUsersBucket = append(ret.mcBucket, "mc_users")
	ret.mcConfigBucket = append(ret.mcBucket, "mc_config")
	ret.webBucket = []string{"web"}
	ret.webUsersBucket = append(ret.webBucket, "web_users")

	ret.userPrefix = "user_"
	ret.configFeaturePrefix = "feature_"

	// Make sure we can access the DB
	ret.getDatabase()

	return ret
}

func (m *Model) getDatabase() {
	var err error
	m.db, err = boltease.Create(c.dir+"/mcman.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

/* Web Server Stuff */
func (m *Model) getAllWebUsers() []string {
	var ret []string
	if err := m.db.OpenDB(); err != nil {
		return ret
	}
	defer m.db.CloseDB()

	userBkts, err := m.db.GetBucketList(m.webUsersBucket)
	if err != nil {
		return ret
	}
	for i := range userBkts {
		var uname string
		userBktPath := append(m.webUsersBucket, userBkts[i])
		if uname, err = m.db.GetValue(userBktPath, "username"); err != nil {
			continue
		}
		ret = append(ret, uname)
	}
	return ret
}

func (m *Model) checkWebUserCreds(username, pw string) error {
	var err error
	if err = m.db.OpenDB(); err != nil {
		return err
	}
	defer m.db.CloseDB()

	userBucketPath := append(m.webUsersBucket, m.userPrefix+username)
	var uPw string
	uPw, err = m.db.GetValue(userBucketPath, "password")
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(uPw), []byte(pw))
}

func (m *Model) updateWebUser(uname, pw string) error {
	var err error
	if err = m.db.OpenDB(); err != nil {
		return err
	}
	defer m.db.CloseDB()

	userBucketPath := append(m.webUsersBucket, m.userPrefix+uname)
	if err = m.db.SetValue(userBucketPath, "username", uname); err != nil {
		return errors.New("Error Updating User (" + uname + "): " + err.Error())
	}
	cryptPw, cryptError := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if cryptError != nil {
		return cryptError
	}
	if err = m.db.SetValue(userBucketPath, "password", string(cryptPw)); err != nil {
		return errors.New("Error Updating User (" + uname + "): " + err.Error())
	}
	return err
}

func (m *Model) mcSaveFeature(opt string, enabled bool) error {
	var err error
	if err = m.db.OpenDB(); err != nil {
		// TODO: Log/output the error
		return err
	}
	defer m.db.CloseDB()

	cfgOption := m.configFeaturePrefix + opt
	if err = m.db.SetBool(m.mcConfigBucket, cfgOption, enabled); err != nil {
		return errors.New("Error Updating Feature (" + opt + "): " + err.Error())
	}
	return err
}

func (m *Model) mcFeatureIsEnabled(opt string) bool {
	ret := false
	var err error
	if err = m.db.OpenDB(); err != nil {
		// TODO: Log/output the error
		return false
	}
	defer m.db.CloseDB()

	cfgOption := m.configFeaturePrefix + opt
	if ret, err = m.db.GetBool(m.mcConfigBucket, cfgOption); err != nil {
		// TODO: Log/output the error
		return false
	}
	return ret
}

/* Minecraft Config Stuff */
func (m *Model) getAllMCUsers() ([]MCUser, error) {
	var ret []MCUser
	var err error
	if err = m.db.OpenDB(); err != nil {
		// TODO: Log/output the error
		return ret, err
	}
	defer m.db.CloseDB()

	userBkts, err := m.db.GetBucketList(m.mcUsersBucket)
	if err != nil {
		// TODO: Log/output error
		return ret, err
	}
	for i := range userBkts {
		userBktPath := append(m.mcUsersBucket, userBkts[i])
		var ld *MCUser
		if ld, err = m.getMCUserFromPath(userBktPath); err != nil {
			continue
		}
		ret = append(ret, *ld)
	}

	return ret, err
}

func (m *Model) getOnlineMCUsers() ([]MCUser, error) {
	var ret []MCUser
	var err error
	if err = m.db.OpenDB(); err != nil {
		// TODO: Log/output the error
		return ret, err
	}
	defer m.db.CloseDB()

	userBkts, err := m.db.GetBucketList(m.mcUsersBucket)
	if err != nil {
		// TODO: Log/output error
		return ret, err
	}
	for i := range userBkts {
		userBktPath := append(m.mcUsersBucket, userBkts[i])
		var ld *MCUser
		if ld, err = m.getMCUserFromPath(userBktPath); err != nil {
			continue
		}
		// Check if ld.LogoutTime < ld.LoginTime
		if ld.LogoutTime.Unix() < ld.LoginTime.Unix() {
			ret = append(ret, *ld)
		}
	}

	return ret, err
}

func (m *Model) getMCUser(nm string) (*MCUser, error) {
	userBktPath := append(m.mcUsersBucket, m.userPrefix+nm)
	return m.getMCUserFromPath(userBktPath)
}

func (m *Model) getMCUserFromPath(pth []string) (*MCUser, error) {
	ret := new(MCUser)
	var err error
	if err = m.db.OpenDB(); err != nil {
		// TODO: Log/output the error
		return nil, err
	}
	defer m.db.CloseDB()

	if ret.Name, err = m.db.GetValue(pth, "name"); err != nil {
		return nil, err
	}
	if ret.IsOp, err = m.db.GetBool(pth, "op"); err != nil {
		return nil, err
	}
	if ret.Home, err = m.db.GetValue(pth, "home"); err != nil {
		return nil, err
	}
	if ret.Porch, err = m.db.GetValue(pth, "porch"); err != nil {
		return nil, err
	}
	var tmpInt int
	if tmpInt, err = m.db.GetInt(pth, "quota"); err != nil {
		return nil, err
	}
	ret.Quota = time.Duration(tmpInt)
	if tmpInt, err = m.db.GetInt(pth, "quotaused"); err != nil {
		return nil, err
	}
	ret.QuotaUsed = time.Duration(tmpInt)
	if ret.LoginTime, err = m.db.GetTimestamp(pth, "logintime"); err != nil {
		return nil, err
	}
	if ret.LogoutTime, err = m.db.GetTimestamp(pth, "logouttime"); err != nil {
		return nil, err
	}
	return ret, err
}

// updateMCUser adds or updates a user
func (m *Model) updateMCUser(u *MCUser) error {
	var err error
	if err = m.db.OpenDB(); err != nil {
		// TODO: Log/output the error
		return err
	}
	defer m.db.CloseDB()

	userBktPath := append(m.mcUsersBucket, m.userPrefix+u.Name)
	if err = m.db.SetValue(userBktPath, "name", u.Name); err != nil {
		return err
	}
	if err = m.db.SetBool(userBktPath, "op", u.IsOp); err != nil {
		return err
	}
	if err = m.db.SetValue(userBktPath, "home", u.Home); err != nil {
		return err
	}
	var tmpInt int
	tmpInt = int(u.Quota)
	if err = m.db.SetInt(userBktPath, "quota", tmpInt); err != nil {
		return err
	}
	tmpInt = int(u.QuotaUsed)
	if err = m.db.SetInt(userBktPath, "quotaused", tmpInt); err != nil {
		return err
	}
	if err = m.db.SetTimestamp(userBktPath, "logintime", u.LoginTime); err != nil {
		return err
	}
	err = m.db.SetTimestamp(userBktPath, "logouttime", u.LogoutTime)
	return err
}
