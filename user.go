package main

import (
	"time"
)

type MCUser struct {
	Name         string
	Index        int
	IsOp         bool
	Home         string
	Porch        string
	ToJSONString func() string
	Quota        time.Duration
	quotaUsed    time.Duration
	loginTime    time.Time
}

func NewMCUser(nm string) *MCUser {
	m := new(MCUser)
	m.Name = nm
	if nm == "" {
		m.Index = -1
	}
	m.IsOp = false
	m.Home = ""
	m.Porch = ""
	m.Quota = 0
	m.quotaUsed = 0
	m.ToJSONString = func() string {
		return "{\"name\":\"" + m.Name + "\",\"home\":\"" + m.Home + "\",\"porch\":\"" + m.Porch + "\",\"quota\":\"" + m.Quota.String() + "\",\"quota_used\":\"" + m.quotaUsed.String() + "\"}"
	}
	return m
}

func (u *MCUser) HasQuota() bool {
	if u.Quota > 0 {
		return u.quotaUsed < u.Quota
	} else {
		return true
	}
}

func (u *MCUser) RemainingQuota() time.Duration {
	if u.Quota > 0 {
		return u.Quota - u.quotaUsed
	} else {
		return 0
	}
}
