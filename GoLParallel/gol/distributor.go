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

type worldSegment struct {
	segment [][]uint8
	start   int
	length  int
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

func makeWorldSegment(p Params, i int, n int, withFringes bool) worldSegment {
	segLength := getSegLength(p, i, n)
	if withFringes {
		segLength = segLength + 2
	}
	return worldSegment{
		segment: makeByteArray(p.ImageWidth, segLength),
		start:   getSegStart(p, i, n),
		length:  segLength,
	}
}

// i is segment number and n is number of threads
func getSegStart(p Params, i int, n int) int {
	if p.ImageHeight%n == 0 {
		return (p.ImageHeight / n) * i
	} else {
		if i != n-1 {
			return ((p.ImageHeight-(p.ImageHeight%n-1))/n - 1) * i
		} else {
			return p.ImageHeight - (p.ImageHeight % (n - 1))
		}
	}
}

func getSegLength(p Params, i int, n int) int {
	if p.ImageHeight%n == 0 {
		return p.ImageHeight / n
	} else {
		if i != n-1 {
			return (p.ImageHeight-(p.ImageHeight%n-1))/n - 1
		} else {
			return p.ImageHeight%n - 1
		}
	}
}

// col is y coordinate and row is x
func splitWorld(p Params, i int, n int, firstWorld [][]uint8) worldSegment {
	// work out how big the segment needs to be, allocate memory
	seg := makeWorldSegment(p, i, n, true)
	if i != 0 && i != n-1 { // if not the first or last segment
		for row := 0; row < seg.length; row++ {
			for col := 0; col < p.ImageWidth; col++ {
				seg.segment[row][col] = firstWorld[seg.start+row-1][col]
			}
		}
	} else if i == 0 { // if first segment
		for col := 0; col < p.ImageWidth; col++ {
			// copy bottom of world in (fringes)
			seg.segment[0][col] = firstWorld[p.ImageHeight-1][col]
		}
		for row := 0; row < seg.length; row++ {
			for col := 0; col < p.ImageWidth; col++ {
				seg.segment[row][col] = firstWorld[seg.start+row][col] // no -1 bcos 1st seg
			}
		}
	} else if i == n-1 { // if last segment
		for col := 0; col < p.ImageWidth; col++ {
			// copy top of world into bottom row of segment (fringes)
			seg.segment[seg.length-1][col] = firstWorld[0][col]
		}
		for row := 0; row < seg.length-1; row++ {
			for col := 0; col < p.ImageWidth; col++ {
				seg.segment[row][col] = firstWorld[seg.start+row-1][col]
			}
		}
	}
	return seg
}

// GOL LOGIC. Has since been moved to worker

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
					// fmt.Println("Continuing execution from turn ", turn)
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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	// PARALLEL GOL
	// load initial world
	//pausing = false
	// pause := <-pausing
	// fmt.Println(pause)
	world := makeByteArray(p.ImageWidth, p.ImageHeight)
	loadFirstWorld(p, world, c)

	// send a cell flipped event for all cells initially alive
	for col := 0; col < p.ImageHeight; col++ {
		for row := 0; row < p.ImageWidth; row++ {
			if world[col][row] == 255 {
				//fmt.Println("yo")
				c.events <- CellFlipped{0, util.Cell{row, col}}
			}
			//fmt.Println("cell flipped done")
		}
	}

	//turn is updated outside of loop

	ticker := time.NewTicker(2 * time.Second)
	turn = 0

	//select {}
	// split world into segments, send each segment to each worker
	for turn < p.Turns {
		go twoSecondTicker(ticker, turn, p, world, c)
		go keyPress(p, world, c, turn, keyPresses)
		// BAD DIVISION
		segmentHeight := p.ImageHeight / p.Threads

		channels := make([]chan [][]byte, p.Threads)
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

			var newWorld [][]byte
			for i := 0; i < p.Threads; i++ {
				newWorld = append(newWorld, <-channels[i]...)
			}

			world = newWorld

			turn++

			c.events <- TurnComplete{turn}

		}
	}

	//do image output shit here
	outputWorld(p, world, c, turn)
	// Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{p.Turns, calculateAliveCells(p, world, c)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}
