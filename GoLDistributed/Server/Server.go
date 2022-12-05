package main

import (
	"GameOfLifeReal/stubs"
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

var ln net.Listener

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

func calculateNextState(req stubs.Request, world [][]byte /*, c distributorChannels*/) [][]byte {
	sum := 0
	segment := req.Segment
	if req.P.Turns == 0 {
		for x := 0; x < req.P.ImageWidth; x++ {
			for y := req.SegStart; y < req.SegEnd; y++ {
				if world[x][y] == 255 {
					segment[x][y] = 255
				}
			}
		}
	} else {
		for x := 0; x < req.P.ImageWidth; x++ {
			for y := req.SegStart; y < req.SegEnd; y++ {
				sum = (int(world[(x+req.P.ImageWidth-1)%req.P.ImageWidth][(y+req.P.ImageHeight-1)%req.P.ImageHeight]) +

					int(world[(x+req.P.ImageWidth)%req.P.ImageWidth][(y+req.P.ImageHeight-1)%req.P.ImageHeight]) +

					int(world[(x+req.P.ImageWidth+1)%req.P.ImageWidth][(y+req.P.ImageHeight-1)%req.P.ImageHeight]) +

					int(world[(x+req.P.ImageWidth-1)%req.P.ImageWidth][(y+req.P.ImageHeight)%req.P.ImageHeight]) +
					int(world[(x+req.P.ImageWidth+1)%req.P.ImageWidth][(y+req.P.ImageHeight)%req.P.ImageHeight]) +
					int(world[(x+req.P.ImageWidth-1)%req.P.ImageWidth][(y+req.P.ImageHeight+1)%req.P.ImageHeight]) +
					int(world[(x+req.P.ImageWidth)%req.P.ImageWidth][(y+req.P.ImageHeight+1)%req.P.ImageHeight]) +
					int(world[(x+req.P.ImageWidth+1)%req.P.ImageWidth][(y+req.P.ImageHeight+1)%req.P.ImageHeight])) / 255
				if world[x][y] == 255 {
					if sum < 2 {
						segment[x][y] = 0
						// c.events <- CellFlipped{turn, util.Cell{x, y}}
					} else if sum == 2 || sum == 3 {
						segment[x][y] = 255
					} else {
						segment[x][y] = 0
						// c.events <- CellFlipped{turn, util.Cell{x, y}}
					}
				} else {
					if sum == 3 {
						segment[x][y] = 255
						// c.events <- CellFlipped{turn, util.Cell{x, y}}
					} else {
						segment[x][y] = 0
					}
				}
			}
		}
	}
	return segment
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
	//SHOULD BE SEGMENTS NOT WORLD BUT PASS WORLD ALSO TO DO COMPUTATION
	world := req.World
	newSegment := req.Segment
	//only calculate next state if the requested turns are greater than 0
	//fmt.Println("the world", world)

	newSegment = calculateNextState(req, world)
	//fmt.Println("the segment", newSegment)
	res.NewSegment = newSegment
	//fmt.Println(res.NewSegment)
	return
}

func (s *GameOfLifeOperations) KillProcess(req stubs.Request, res stubs.Response) (err error) {
	ln.Close()
	return
}

func main() {
	pAddr := flag.String("port", "8031", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	ln = listener
	defer listener.Close()
	rpc.Accept(listener)

}
