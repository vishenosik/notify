package config

type Environment string

func (e Environment) IsDev() bool {
	return e == "dev"
}

func (e Environment) String() string {
	return string(e)
}
