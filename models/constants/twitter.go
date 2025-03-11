package constants

import "time"

type TwitterAccount struct {
	ID         string
	Name       string
	Symbol     string
	LastUpdate time.Time
}

func GetTwitterAccounts() []TwitterAccount {
	var twitterAccounts []TwitterAccount
	twitterAccounts = append(twitterAccounts, TwitterAccount{ID: "1", Name: "iEx_ec", Symbol: "RLC", LastUpdate: time.Now().UTC()})
	return twitterAccounts
}
