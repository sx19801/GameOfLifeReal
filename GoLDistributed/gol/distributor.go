package gol

import (
	"GameOfLifeReal/stubs"
	"GameOfLifeReal/util"
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"sync"
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

func outputWorld(p Params, world [][]byte, c distributorChannels, turn int) {
	c.ioCommand <- 0
	c.ioFilename <- strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(turn)
	for i := 0; i < p.ImageWidth; i++ {
		for j := 0; j < p.ImageHeight; j++ {
			c.ioOutput <- world[i][j]
		}
	}
}

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

var wg sync.WaitGroup
var pausing bool
var running bool

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, key <-chan rune) {

	// TODO: Create a 2D slice to store the world.
	firstWorld := makeByteArray(p)
	// Get initial world as input from io channel and populate
	loadFirstWorld(p, firstWorld, c)
	// TODO: Execute all turns of the Game of Life.
	//finalWorld := makeByteArray(p)
	//running := true
	// ticker := time.NewTicker(2 * time.Second)
	//rpc call shit
	server := "127.0.0.1:8030"
	flag.Parse()
	fmt.Println("Server: ", server)
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()
	turn := 0
	// var segment [][]byte
	pausing = false

	request := stubs.Request{World: firstWorld, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
	response := new(stubs.Response)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:

				client.Call(stubs.AliveCells, request, response)

				if response.CurrentTurn != 0 {
					a := len(calculateAliveCells(p, response.GlobalWorld, c))
					c.events <- AliveCellsCount{response.CurrentTurn, a}
				}
			case <-key:
				if <-key == 's' {
					outputWorld(p, response.NewWorld, c, turn)
				} else if <-key == 'q' {
					fmt.Println("closing client")
					client.Close()
					running = false
					c.events <- StateChange{turn, Quitting}
				} else if <-key == 'k' {

					client.Call(stubs.KillServer, request, response)
					outputWorld(p, response.NewWorld, c, turn)
					//send kill request down channel to server
					client.Close()
					running = false
				} else if <-key == 'p' {
					if pausing {
						pausing = false
						wg.Done()
						fmt.Println("Continuing")
						break
					}
					wg.Add(1)
					outputWorld(p, response.NewWorld, c, turn)

					pausing = true
				}
				// fmt.Println(response.CurrentTurn)
				//fmt.Println(len(response.GlobalWorld))
				//fmt.Println(len(calculateAliveCells(p, response.GlobalWorld, c)))
				// a := len(calculateAliveCells(p, response.GlobalWorld, c))
				// c.events <- AliveCellsCount{response.CurrentTurn, a}
				//case <-keyboardpresses:

			}
		}
	}()
	fmt.Println("before client call in distributor")
	client.Call(stubs.BrokerHandler, request, response)
	fmt.Println("after client call in distributor")

	// send request
	//extract
	//finalWorld = gameOfLife(p, firstWorld, c)
	// TODO: Report the final state using FinalTurnCompleteEvent.
	//fmt.Println(len(response.NewWorld))
	outputWorld(p, response.NewWorld, c, turn)
	//fmt.Println("after output world")

	c.events <- FinalTurnComplete{p.Turns, calculateAliveCells(p, response.NewWorld, c)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	//fmt.Println("before close")
	close(c.events)
}
