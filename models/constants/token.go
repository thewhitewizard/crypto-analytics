package constants

type CrytoWatch struct {
	CryptoId int
	Symbol   string
	Gecko    string
	Handle   string
	Desc     string
}

func GetCrytoWatch() []CrytoWatch {
	var cryptocurrencies []CrytoWatch
	cryptocurrencies = append(cryptocurrencies, CrytoWatch{Gecko: "iexec-rlc", Symbol: "RLC", Desc: "iExec RLC (RLC)", Handle: "IExecRLC", CryptoId: 1637})
	cryptocurrencies = append(cryptocurrencies, CrytoWatch{Gecko: "pha", Symbol: "PHA", Desc: "Phala Network (PHA)", Handle: "PhalaNetwork", CryptoId: 6841})
	cryptocurrencies = append(cryptocurrencies, CrytoWatch{Gecko: "secret", Symbol: "SCRT", Desc: "Secret Network (SCRT)", Handle: "secretnetwork", CryptoId: 5604})
	//cryptocurrencies = append(cryptocurrencies, CrytoWatch{Gecko: "secret", Symbol: "GLM", Desc: "Golem", Handle: "secretnetwork", CryptoId: 1455})

	return cryptocurrencies
}

//	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum,iexec-rlc,secret,pha,akash-network,golem&vs_currencies=usd"
