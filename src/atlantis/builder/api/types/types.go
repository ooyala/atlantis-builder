package types

const (
	StatusInit     = "INIT"
	StatusDone     = "DONE"
	StatusError    = "ERROR"
	StatusBuilding = "Building..."
	StatusBooting  = "Booting..."
)

type Build struct {
	ID      string
	URL     string
	Sha     string
	RelPath string
	Status  string
	Error   interface{}
}

type Boot struct {
	Status string
	Error  interface{}
}
