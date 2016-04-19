package models

import (
	"time"
)

type CntAnalysis struct {
	Src       string
	Timestamp time.Time
	Action    string
	Repo      string
	User      string
}

type CntRepo struct {
	User string
	Repo string
	Tag  string
}

type CntUser struct {
	Username string
	Password string
}

type ACLEntry struct {
	Match MatchConditions `yaml:"match"`
}

type MatchConditions struct {
	Account string `yaml:"account,omitempty" json:"account,omitempty"`
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
}
