package gol

import (
	"GameOfLifeReal/util"
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
	c.ioFilename <- "16x16"
	for i := 0; i < p.ImageWidth; i++ {
		for j := 0; j < p.ImageHeight; j++ {
			firstWorld[i][j] = <-c.ioInput
		}
	}
}

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

func gameOfLife(p Params, world [][]byte, c distributorChannels) [][]byte {
	for turn := 0; turn < p.Turns; turn++ {
		world = calculateNextState(p, world, c, turn)
		c.events <- TurnComplete{turn}
	}
	return world
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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	firstWorld := makeByteArray(p)
	// Get initial world as input from io channel and populate
	loadFirstWorld(p, firstWorld, c)
	// TODO: Execute all turns of the Game of Life.
	finalWorld := makeByteArray(p)
	finalWorld = gameOfLife(p, firstWorld, c)
	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{p.Turns, calculateAliveCells(p, finalWorld, c)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
