package stubs

var GolHandler = "GameOfLifeOperations.ProcessGameOfLife"
var KillServer = "GameOfLifeOperations.KillProcess"

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type Response struct {
	NewWorld    [][]byte
	CurrentTurn int
}

type Request struct {
	World [][]byte
	P     Params
}
