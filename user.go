package main

import (
	"time"
)

type MCUser struct {
	Name       string
	Index      int
	IsOp       bool
	Home       string
	Porch      string
	Quota      time.Duration
	QuotaUsed  time.Duration
	LoginTime  time.Time
	LogoutTime time.Time
}

func NewMCUser(nm string) *MCUser {
	if nm == "" {
		return nil
	}
	m := new(MCUser)
	m.Name = nm
	return m
}

func (u *MCUser) HasQuota() bool {
	if u.Quota > 0 {
		return u.QuotaUsed < u.Quota
	} else {
		return true
	}
}

func (u *MCUser) RemainingQuota() time.Duration {
	if u.Quota > 0 {
		return u.Quota - u.QuotaUsed
	} else {
		return 0
	}
}
