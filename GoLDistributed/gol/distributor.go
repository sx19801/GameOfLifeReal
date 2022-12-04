package gol

import (
	"GameOfLifeReal/stubs"
	"GameOfLifeReal/util"
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"time"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func makeByteArray(p Params) [][]byte {
	newArray := make([][]byte, p.ImageWidth)
	for i := 0; i < p.ImageWidth; i++ {
		newArray[i] = make([]byte, p.ImageHeight)
	}
	return newArray
}

func loadFirstWorld(p Params, firstWorld [][]byte, c distributorChannels) {
	c.ioCommand <- 1
	c.ioFilename <- strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth)
	for i := 0; i < p.ImageWidth; i++ {
		for j := 0; j < p.ImageHeight; j++ {
			firstWorld[i][j] = <-c.ioInput
		}
	}
}

/*
func calculateNextState(p Params, world [][]byte, c distributorChannels, turn int) [][]byte {
	sum := 0
	newWorld := make([][]byte, p.ImageWidth)
	for i := 0; i < p.ImageWidth; i++ {
		newWorld[i] = make([]byte, p.ImageHeight)
	}
	for x := 0; x < p.ImageWidth; x++ {
		for y := 0; y < p.ImageHeight; y++ {
			sum = (int(world[(y+p.ImageHeight-1)%p.ImageHeight][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world[(y+p.ImageHeight-1)%p.ImageHeight][(x+p.ImageWidth)%p.ImageWidth]) +
				int(world[(y+p.ImageHeight-1)%p.ImageHeight][(x+p.ImageWidth+1)%p.ImageWidth]) +
				int(world[(y+p.ImageHeight)%p.ImageHeight][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world[(y+p.ImageHeight)%p.ImageHeight][(x+p.ImageWidth+1)%p.ImageWidth]) +
				int(world[(y+p.ImageHeight+1)%p.ImageHeight][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world[(y+p.ImageHeight+1)%p.ImageHeight][(x+p.ImageWidth)%p.ImageWidth]) +
				int(world[(y+p.ImageHeight+1)%p.ImageHeight][(x+p.ImageWidth+1)%p.ImageWidth])) / 255
			if world[y][x] == 255 {
				if sum < 2 {
					newWorld[y][x] = 0
					c.events <- CellFlipped{turn, util.Cell{x, y}}
				} else if sum == 2 || sum == 3 {
					newWorld[y][x] = 255
				} else {
					newWorld[y][x] = 0
					c.events <- CellFlipped{turn, util.Cell{x, y}}
				}
			} else {
				if sum == 3 {
					newWorld[y][x] = 255
					c.events <- CellFlipped{turn, util.Cell{x, y}}
				} else {
					newWorld[y][x] = 0
				}
			}
		}
	}
	return newWorld
}
*/
/*func gameOfLife(p Params, world [][]byte, c distributorChannels) [][]byte {
	for turn := 0; turn < p.Turns; turn++ {
		world = calculateNextState(p, world, c, turn)
		c.events <- TurnComplete{turn}
	}
	return world
}
*/

func calculateAliveCells(p Params, world [][]byte, c distributorChannels) []util.Cell {
	aliveCells := make([]util.Cell, 0)
	for x := 0; x < p.ImageWidth; x++ {
		for y := 0; y < p.ImageHeight; y++ {
			if world[y][x] == 255 {
				aliveCells = append(aliveCells, util.Cell{x, y})
			}
		}
	}
	return aliveCells
}

// func makeCall(client *rpc.Client, world [][]byte, p stubs.Params) [][]byte {
// 	request := stubs.Request{World: world, P: p}
// 	response := new(stubs.Response)
// 	client.Call(stubs.GolHandler, request, response)
// 	return response.NewWorld
// }

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	firstWorld := makeByteArray(p)
	// Get initial world as input from io channel and populate
	loadFirstWorld(p, firstWorld, c)
	// TODO: Execute all turns of the Game of Life.
	//finalWorld := makeByteArray(p)
	running := true
	ticker := time.NewTicker(2 * time.Second)
	//rpc call shit
	server := "127.0.0.1:8030"
	flag.Parse()
	fmt.Println("Server: ", server)
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()
	request := stubs.Request{World: firstWorld, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
	response := new(stubs.Response)

	call := client.Go(stubs.GolHandler, request, response, nil)
	//fmt.Println("p.turns", p.Turns)
	for running {
		// 	// 	//request := stubs.Request{World: response.NewWorld, P: stubs.Params{p.ImageHeight, p.ImageWidth, p.Threads, response.CurrentTurn}}
		// 	// 	response := new(stubs.Response)
		select {
		case <-ticker.C:
			fmt.Println("yo")
			//<-call.Done
		// 		client.Call(stubs.GolHandler, stubs.Request{World: response.NewWorld, P: stubs.Params{p.ImageHeight, p.ImageWidth, p.Threads, response.CurrentTurn}}, response)

		// 		<-call.Done
		// 		// 		//fmt.Println("before call", response.NewWorld)
		// 		// 		//fmt.Println("before call", <-call.Done)
		// 		// 		client.Call(stubs.GolHandler, stubs.Request{World: response.NewWorld, P: stubs.Params{p.ImageHeight, p.ImageWidth, p.Threads, p.Turns}}, response)
		// 		// 		//fmt.Println(d)
		// 		// 		fmt.Println("after call ", response.CurrentTurn)

		// 		// 		//c.events <- AliveCellsCount{response.CurrentTurn, len(calculateAliveCells(p, response.NewWorld, c))}

		case <-call.Done:
			running = false
		}
	}

	// send request
	//extract
	//finalWorld = gameOfLife(p, firstWorld, c)
	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{p.Turns, calculateAliveCells(p, response.NewWorld, c)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
