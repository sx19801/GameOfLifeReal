package main

import (
	"GameOfLifeReal/stubs"
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

func makeByteArray(p stubs.Params) [][]byte {
	newArray := make([][]byte, p.ImageWidth)
	for i := 0; i < p.ImageWidth; i++ {
		newArray[i] = make([]byte, p.ImageHeight)
	}
	return newArray
}

// func loadFirstWorld(p Params, firstWorld [][]byte, c distributorChannels) {
// 	c.ioCommand <- 1
// 	c.ioFilename <- strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth)
// 	for i := 0; i < p.ImageWidth; i++ {
// 		for j := 0; j < p.ImageHeight; j++ {
// 			firstWorld[i][j] = <-c.ioInput
// 		}
// 	}
// }

func calculateNextState(p stubs.Params, world [][]byte /*, c distributorChannels*/) [][]byte {
	sum := 0
	newWorld := makeByteArray(p)
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
					// c.events <- CellFlipped{turn, util.Cell{x, y}}
				} else if sum == 2 || sum == 3 {
					newWorld[y][x] = 255
				} else {
					newWorld[y][x] = 0
					// c.events <- CellFlipped{turn, util.Cell{x, y}}
				}
			} else {
				if sum == 3 {
					newWorld[y][x] = 255
					// c.events <- CellFlipped{turn, util.Cell{x, y}}
				} else {
					newWorld[y][x] = 0
				}
			}
		}
	}
	return newWorld
}

/*
	func calculateAliveCells(p stubs.Params, world [][]byte) []util.Cell {
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
type GameOfLifeOperations struct{}

func (s *GameOfLifeOperations) ProcessGameOfLife(req stubs.Request, res *stubs.Response) (err error) {

	newWorld := req.World

	//only calculate next state if the requested turns are greater than 0
	if req.P.Turns != 0 {
		newWorld = calculateNextState(req.P, newWorld)
	}
	res.NewWorld = newWorld
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)

}
