package interfaces

type ConfigManagement struct {
	Enabled bool
	Api     string `toml:"api_location"`
}
