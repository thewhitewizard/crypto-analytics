package constants

type CrytoWatch struct {
	CryptoId int
	Symbol   string
	Handle   string
	Desc     string
}

func GetCrytoWatch() []CrytoWatch {
	var cryptocurrencies []CrytoWatch
	cryptocurrencies = append(cryptocurrencies, CrytoWatch{Symbol: "RLC", Desc: "iExec RLC (RLC)", Handle: "IExecRLC", CryptoId: 1637})
	cryptocurrencies = append(cryptocurrencies, CrytoWatch{Symbol: "PHA", Desc: "Phala Network (PHA)", Handle: "PhalaNetwork", CryptoId: 6841})
	cryptocurrencies = append(cryptocurrencies, CrytoWatch{Symbol: "SCRT", Desc: "Secret Network (SCRT)", Handle: "secretnetwork", CryptoId: 5604})

	return cryptocurrencies
}
