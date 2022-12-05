package stubs

var GolHandler = "GameOfLifeOperations.ProcessGameOfLife"
var KillServer = "GameOfLifeOperations.KillProcess"
var BrokerHandler = "GameOfLifeOperations.BrokerProcessGol"

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type Response struct {
	NewWorld    [][]byte
	NewSegment  [][]byte
	CurrentTurn int
}

type Request struct {
	World    [][]byte
	Segment  [][]byte
	P        Params
	SegStart int
	SegEnd   int
}
