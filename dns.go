package essos

type DNS struct {
	Enabled bool
	Path    string
	Api     string `toml:"api-location"`
	Etcd    string `toml:"etcd-address"`
	Domain  string
}
