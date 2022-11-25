package gol

import "GameOfLifeReal/util"

type workerChannels struct {
	in  chan worldSegment
	out chan worldSegment
}

// same as calculateNextState
func calculateNextStateOfSegment(p Params, world worldSegment) worldSegment {
	sum := 0
	// make smaller segment to return processed section without the fringes
	newSegment := worldSegment{
		segment: makeByteArray(p.ImageWidth, world.length-2),
		start:   world.start,
		length:  world.length - 2,
	}

	for y := 1; y < len(world.segment)-1; y++ {
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
					newSegment.segment[y][x] = 0
				} else if sum == 2 || sum == 3 {
					newSegment.segment[y][x] = 255
				} else {
					newSegment.segment[y][x] = 0
				}
			} else {
				if sum == 3 {
					newSegment.segment[y][x] = 255
				} else {
					newSegment.segment[y][x] = 0
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

func work(w workerChannels, d distributorChannels, p Params) {
	firstSegment := <-w.in
	nextSegment := calculateNextStateOfSegment(p, firstSegment)
	w.out <- nextSegment
}
