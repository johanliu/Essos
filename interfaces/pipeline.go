package interfaces

type Pipeline struct {
	Enabled bool
	API     string `toml:"api_location"`
	IP      string
	Port    string
}
