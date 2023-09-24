package testing

type scenario int

const (
	identical scenario = iota
	scaled
	rotated
	mirrored
	moved
	background
	motive
	part
	mixed
)

func (s scenario) string() string {
	names := [...]string{"identical", "scaled", "rotated", "mirrored", "moved", "background", "motive", "part"}
	if s < identical || s > mixed {
		return "nil"
	}
	return names[s]
}

func runScenario() {

}
