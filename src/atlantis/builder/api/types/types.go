package types

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
