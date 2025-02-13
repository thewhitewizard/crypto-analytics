package constants

type TwitterAccount struct {
	ID   string
	Name string
}

func GetTwitterAccounts() []TwitterAccount {
	var twitterAccounts []TwitterAccount
	twitterAccounts = append(twitterAccounts, TwitterAccount{ID: "1", Name: "iEx_ec"})
	return twitterAccounts
}
