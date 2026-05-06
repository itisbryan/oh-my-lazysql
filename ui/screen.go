package ui

type Screen int

const (
	ScreenConnectionList Screen = iota
	ScreenConnectionForm
	ScreenHome
)

type ScreenChangeMsg struct {
	Screen Screen
	Data   any
}

type ConnectionSelectedMsg struct {
	Connection any
}

type ErrorMsg struct {
	Err error
}