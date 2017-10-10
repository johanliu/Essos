package interfaces

type DNS struct {
	Enabled bool
	Api     string `toml:"api_location"`
	Etcd    string `toml:"etcd_address"`
	Domain  string
}
