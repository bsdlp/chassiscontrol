package configuration

type Object struct {
	ServerHostPort string
	Targets        map[string]IPMIHostConfig
}

type IPMIHostConfig struct {
	Username string
	Password string
	Address  string
}
