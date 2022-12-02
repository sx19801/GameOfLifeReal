package gol

import (
	"strconv"
	"sync"
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

/*
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
*/
// 	defer ticker.Stop()
// 	if turn != 0 {
// 		for _ = range ticker.C {
// 			fmt.Println(turn)
// 			mutex.Lock()
// 			c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world, c))}
// 			mutex.Unlock()
// 		}
// 	}

// }
func twoSecondTicker(ticker *time.Ticker, turn int, p Params, world [][]byte, c distributorChannels) {
	// if turn%100 == 0 {
	// 	fmt.Println("inside ticker ", turn)
	// }

	for {
		select {
		case <-ticker.C:

			//fmt.Println("ticker ", turn)

			// if len(calculateAliveCells(p, world, c))%5 == 0 {
			// 	fmt.Println(len(calculateAliveCells(p, world, c)))
			// }
			if turn != 0 {
				c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world, c))}
			}

		default:
			// if turn%100 == 0 {
			// 	fmt.Println("default ", turn)
			// }
			return
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, mutex sync.Mutex) {
	// SERIAL GOL
	// Create a 2D slice to store the world.
	/*
		firstWorld := makeByteArray(p)
		// Get initial world as input from io channel and populate
		loadFirstWorld(p, firstWorld, c)
		// Execute all turns of the Game of Life.
		finalWorld := makeByteArray(p)
		finalWorld = gameOfLife(p, firstWorld, c)
	*/

	// PARALLEL GOL
	// workerChannels := workerChannels{
	// 	in:  make(chan worldSegment),
	// 	out: make(chan worldSegment),
	// }
	// load initial world
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
	turn := 0

	//select {}
	// split world into segments, send each segment to each worker
	for ; turn < p.Turns; turn++ {
		go twoSecondTicker(ticker, turn, p, world, c)
		// BAD DIVISION
		segmentHeight := p.ImageHeight / p.Threads

		channels := make([]chan [][]byte, p.Threads)
		for i := range channels {
			channels[i] = make(chan [][]byte)
		}

		for i := 0; i < p.Threads; i++ {
			if i == p.Threads-1 {
				go calculateNextState(p, world, c, turn, segmentHeight*i, p.ImageHeight, channels[i])
			} else {
				go calculateNextState(p, world, c, turn, segmentHeight*i, segmentHeight*(i+1), channels[i])
				//fmt.Println("next state calculated")
			}

		}

		var newWorld [][]byte
		for i := 0; i < p.Threads; i++ {
			newWorld = append(newWorld, <-channels[i]...)
		}

		// yes

		// for recieved := 0; recieved < p.Threads; recieved++ {
		// 	processedSeg := <-workerChannels.out
		// 	for row := 0; row < processedSeg.length; row++ {
		// 		for col := 0; col < p.ImageWidth; col++ {
		// 			world[processedSeg.start+row][col] = processedSeg.segment[row][col]
		// 		}
		// 	}
		// }

		//mutex.Lock()

		//mutex.Unlock()
		//done := make(chan bool)

		// select {
		// case <-ticker.C:
		// 	if turn != 0 {
		// 		c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world, c))}
		// 	}
		// case <-done:
		// 	if turn%100 == 0 {
		// 		fmt.Println("default", turn)
		// 	}
		// }

		// if turn%100 == 0 {
		// 	fmt.Println("distributor", turn)
		// }

		world = newWorld
		//fmt.Println("before turn complete")
		c.events <- TurnComplete{turn}

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

//uv
