package stubs

var GolHandler = "GameOfLifeOperations.ProcessGameOfLife"
var KillServer = "GameOfLifeOperations.KillProcess"
var BrokerHandler = "GameOfLifeOperations.BrokerProcessGol"
var AliveCells = "GameOfLifeOperations.AliveCellsTicker"

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type Response struct {
	NewWorld    [][]byte
	GlobalWorld [][]byte
	NewSegment  [][]byte
	CurrentTurn int
}

type Request struct {
	World [][]byte

	P        Params
	SegStart int
	SegEnd   int
}
