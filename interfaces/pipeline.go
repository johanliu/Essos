package interfaces

type Pipeline struct {
	Enabled bool
	Type    string
	API     string `toml:"api_location"`
	IP      string
	Port    string
}
