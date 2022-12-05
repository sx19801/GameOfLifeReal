package stubs

var GolHandler = "GameOfLifeOperations.ProcessGameOfLife"

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