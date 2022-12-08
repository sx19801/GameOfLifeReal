package gol

import (
	"fmt"
	"strconv"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

var turn int
var pausing bool

func makeByteArray(x int, height int) [][]byte {
	newArray := make([][]byte, height)
	for i := 0; i < height; i++ {
		newArray[i] = make([]byte, x)
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

func twoSecondTicker(ticker *time.Ticker, turn int, p Params, world [][]byte, c distributorChannels) {
	select {
	case <-ticker.C:
		c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world, c))}
	default:
	}

}

func keyPress(p Params, world [][]byte, c distributorChannels, turn int, keyPresses <-chan rune) {
	for {
		select {
		case <-keyPresses:
			key := <-keyPresses
			if key == 'p' {
				if pausing {
					pausing = false

					c.events <- StateChange{CompletedTurns: turn, NewState: 1}
					break
				}
				pausing = true
				c.events <- StateChange{CompletedTurns: turn, NewState: 0}

			} else if key == 'q' {
				fmt.Println("Printing PGM of current turn ")
				outputWorld(p, world, c, turn)
				c.events <- StateChange{CompletedTurns: turn, NewState: 2}
				time.Sleep(1200 * time.Millisecond)
				c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world, c)}

			} else if key == 's' {
				fmt.Println("Printing PGM of current turn ")
				outputWorld(p, world, c, turn)
			}
		default:
			return
		}
	}
}

func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {

	world := makeByteArray(p.ImageWidth, p.ImageHeight)
	loadFirstWorld(p, world, c)

	// send a cell flipped event for all cells initially alive
	for col := 0; col < p.ImageHeight; col++ {
		for row := 0; row < p.ImageWidth; row++ {
			if world[col][row] == 255 {
				c.events <- CellFlipped{0, util.Cell{row, col}}
			}

		}
	}

	ticker := time.NewTicker(2 * time.Second)
	turn = 0
	segmentHeight := p.ImageHeight / p.Threads

	for turn < p.Turns {
		go twoSecondTicker(ticker, turn, p, world, c)
		go keyPress(p, world, c, turn, keyPresses)

		channels := make([]chan [][]byte, p.Threads)
		var newWorld [][]byte
		for i := range channels {
			channels[i] = make(chan [][]byte)
		}

		if !pausing {
			for i := 0; i < p.Threads; i++ {
				if i == p.Threads-1 {
					go calculateNextState(p, world, c, turn, segmentHeight*i, p.ImageHeight, channels[i])
				} else {
					go calculateNextState(p, world, c, turn, segmentHeight*i, segmentHeight*(i+1), channels[i])
				}
			}

			for i := 0; i < p.Threads; i++ {
				newWorld = append(newWorld, <-channels[i]...)
			}

			world = newWorld
			turn++
			c.events <- TurnComplete{turn}

		}
	}

	outputWorld(p, world, c, turn)
	c.events <- FinalTurnComplete{p.Turns, calculateAliveCells(p, world, c)}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	close(c.events)

}
