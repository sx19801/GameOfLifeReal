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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, key <-chan rune) {

	firstWorld := makeByteArray(p)

	loadFirstWorld(p, firstWorld, c)

	ticker := time.NewTicker(2 * time.Second)

	server := "127.0.0.1:8030"
	flag.Parse()
	fmt.Println("Server: ", server)
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()
	turn := 0
	running := true
	pausing = false
	request := stubs.Request{World: firstWorld, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
	response := new(stubs.Response)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.events <- AliveCellsCount{turn, len(calculateAliveCells(p, response.NewWorld, c))}
			}
		}
	}()

	go func() {
		for running {
			select {
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
					client.Close()
					running = false
				} else if <-key == 'p' {
					if pausing {
						pausing = false
						wg.Done()

						break
					}
					wg.Add(1)
					outputWorld(p, response.NewWorld, c, turn)
					fmt.Println("Pausing")
					pausing = true
				}
			}
		}
	}()

	for running {
		if p.Turns == 0 {
			client.Call(stubs.GolHandler, request, response)
			running = false
			fmt.Println("after wait")
		} else {
			for turn < p.Turns {
				wg.Wait()
				client.Call(stubs.GolHandler, request, response)
				request.World = response.NewWorld
				turn++
				if !running {
					break
				}
			}
			running = false
		}
	}

	outputWorld(p, response.NewWorld, c, turn)

	c.events <- FinalTurnComplete{p.Turns, calculateAliveCells(p, response.NewWorld, c)}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	close(c.events)
}
