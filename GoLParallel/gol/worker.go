package gol

import "GameOfLifeReal/util"

type workerChannels struct {
	in  chan worldSegment
	out chan worldSegment
}

// same as calculateNextState
func calculateNextStateOfSegmentWithFringes(p Params, world worldSegment, c distributorChannels, turn int) worldSegment {
	sum := 0
	// make smaller segment to return processed section without the fringes if more than one thread
	newSegment := worldSegment{
		segment: makeByteArray(p.ImageWidth, world.length-2),
		start:   world.start,
		length:  world.length - 2,
	}

	for y := 1; y < world.length-2; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			sum = (int(world.segment[y-1][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world.segment[y-1][(x+p.ImageWidth)%p.ImageWidth]) +
				int(world.segment[y-1][(x+p.ImageWidth+1)%p.ImageWidth]) +
				int(world.segment[y][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world.segment[y][(x+p.ImageWidth+1)%p.ImageWidth]) +
				int(world.segment[y+1][(x+p.ImageWidth-1)%p.ImageWidth]) +
				int(world.segment[y+1][(x+p.ImageWidth)%p.ImageWidth]) +
				int(world.segment[y+1][(x+p.ImageWidth+1)%p.ImageWidth])) / 255
			if world.segment[y][x] == 255 {
				if sum < 2 {
					newSegment.segment[y-1][x] = 0
					c.events <- CellFlipped{turn, util.Cell{x, world.start + y - 1}}
				} else if sum == 2 || sum == 3 {
					newSegment.segment[y-1][x] = 255
					c.events <- CellFlipped{turn, util.Cell{x, world.start + y - 1}}
				} else {
					newSegment.segment[y-1][x] = 0
					c.events <- CellFlipped{turn, util.Cell{x, world.start + y - 1}}
				}
			} else {
				if sum == 3 {
					newSegment.segment[y-1][x] = 255
					c.events <- CellFlipped{turn, util.Cell{x, world.start + y - 1}}
				} else {
					newSegment.segment[y-1][x] = 0
				}
			}
		}
	}
	return newSegment
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

func work(w workerChannels, d distributorChannels, p Params, turn int) {
	firstSegment := <-w.in
	if p.Threads == 1 {
		w.out <- calculateNextState(p, firstSegment.segment, d, turn)
	} else {
		w.out <- calculateNextStateOfSegmentWithFringes(p, firstSegment, d, turn)
	}
}

///utfytdkut
