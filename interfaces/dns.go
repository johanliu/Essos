package interfaces

type DNS struct {
	Enabled bool
	Path    string `toml:"library_path"`
	Api     string `toml:"api_location"`
	Etcd    string `toml:"etcd_address"`
	Domain  string
}
