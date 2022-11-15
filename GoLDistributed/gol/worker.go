package gol

import "GameOfLifeReal/util"

type workerChannels struct {
	in 		<-chan [][]uint8
	out 	chan<- [][]uint8
	id 		int
}

// same as calculateNextState
func calculateNextStateOfSegment(p Params, world [][]byte, c distributorChannels, turn int) [][]byte {
	sum := 0
	newSegment := make([][]byte, len(world)-2)
	for i := 0; i < p.ImageWidth; i++ {
		newSegment[i] = make([]byte, p.ImageWidth)
	}
	for y := 1; y < len(world)-1; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			sum = (int(world[(y-1][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world[y-1][(x+p.ImageWidth)%p.ImageWidth]) +
				int(world[y-1][(x+p.ImageWidth+1)%p.ImageWidth]) +
				int(world[y][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world[y][(x+p.ImageWidth+1)%p.ImageWidth]) +
				int(world[y+1][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world[y+1%p.ImageHeight][(x+p.ImageWidth)%p.ImageWidth]) +
				int(world[y+1][(x+p.ImageWidth+1)%p.ImageWidth])) / 255
			if world[y][x] == 255 {
				if sum < 2 {
					newSegment[y][x] = 0
					c.events <- CellFlipped{turn, util.Cell{x, y}}
				} else if sum == 2 || sum == 3 {
					newSegment[y][x] = 255
				} else {
					newSegment[y][x] = 0
					c.events <- CellFlipped{turn, util.Cell{x, y}}
				}
			} else {
				if sum == 3 {
					newSegment[y][x] = 255
					c.events <- CellFlipped{turn, util.Cell{x, y}}
				} else {
					newSegment[y][x] = 0
				}
			}
		}
	}
	return newSegment
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
 func work(c workerChannels) {
	 firstSegment := <-c.in
	 nextSegment := calculateNextStateOfSegment()

 }
