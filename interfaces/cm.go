package interfaces

type ConfigManagement struct {
	Enabled bool
	Path    string `toml:"library_path"`
	Api     string `toml:"api_location"`
}
