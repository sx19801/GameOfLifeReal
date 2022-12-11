package main

import (
	"GameOfLifeReal/stubs"
	"flag"
	"fmt"
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
	//segment := req.Segment
	fmt.Println("END: ", req.SegEnd, " Start: ", req.SegStart)
	segment := make([][]byte, req.SegEnd-req.SegStart)
	for i := 0; i < req.SegEnd-req.SegStart; i++ {
		segment[i] = make([]byte, req.P.ImageWidth)
	}

	if req.P.Turns == 0 {
		for x := 0; x < req.P.ImageWidth; x++ {
			for y := req.SegStart; y < req.SegEnd; y++ {
				if world[x][y] == 255 {
					segment[x][y] = 255
				}
			}
		}
	} else {
		for y := req.SegStart; y < req.SegEnd; y++ {
			for x := 0; x < req.P.ImageWidth; x++ {

				sum = (int(world[(y+req.P.ImageHeight-1)%req.P.ImageHeight][(x+req.P.ImageWidth-1)%req.P.ImageWidth]) +

					int(world[(y+req.P.ImageHeight-1)%req.P.ImageHeight][(x+req.P.ImageWidth)%req.P.ImageWidth]) +

					int(world[(y+req.P.ImageHeight-1)%req.P.ImageHeight][(x+req.P.ImageWidth+1)%req.P.ImageWidth]) +

					int(world[(y+req.P.ImageHeight)%req.P.ImageHeight][(x+req.P.ImageWidth-1)%req.P.ImageWidth]) +
					int(world[(y+req.P.ImageHeight)%req.P.ImageHeight][(x+req.P.ImageWidth+1)%req.P.ImageWidth]) +
					int(world[(y+req.P.ImageHeight+1)%req.P.ImageHeight][(x+req.P.ImageWidth-1)%req.P.ImageWidth]) +
					int(world[(y+req.P.ImageHeight+1)%req.P.ImageHeight][(x+req.P.ImageWidth)%req.P.ImageWidth]) +
					int(world[(y+req.P.ImageHeight+1)%req.P.ImageHeight][(x+req.P.ImageWidth+1)%req.P.ImageWidth])) / 255
				if world[y][x] == 255 {
					if sum < 2 {
						segment[y-req.SegStart][x] = 0
						// c.events <- CellFlipped{turn, util.Cell{x, y}}
					} else if sum == 2 || sum == 3 {
						segment[y-req.SegStart][x] = 255
					} else {
						segment[y-req.SegStart][x] = 0
						// c.events <- CellFlipped{turn, util.Cell{x, y}}
					}
				} else {
					if sum == 3 {
						segment[y-req.SegStart][x] = 255
						// c.events <- CellFlipped{turn, util.Cell{x, y}}
					} else {
						segment[y-req.SegStart][x] = world[y][x]
					}
				}
			}
		}
	}
	fmt.Println("WASSUP :", len(segment))
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

var ports = make([]int, 16)
var i int

func makePorts(p stubs.Params) {
	j := 0
	for j < p.Threads {
		k := 30 + j
		ports[j] = k
		j++
	}
}

type GameOfLifeOperations struct{}

func (s *GameOfLifeOperations) ProcessGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	//SHOULD BE SEGMENTS NOT WORLD BUT PASS WORLD ALSO TO DO COMPUTATION
	makePorts(req.P)
	world := req.World
	fmt.Println(world)
	//only calculate next state if the requested turns are greater than 0
	//fmt.Println("the world", world)

	newSegment := calculateNextState(req, world)
	fmt.Println(newSegment)
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
	// +strconv.Itoa(i)
	pAddr := flag.String("ip", "172.25.106.102:8031", "Ip and Port to listen on")
	//brokerAddr := flag.String("port", "8030", "Port to listen on")
	//client, _ := rpc.Dial("tcp", *brokerAddr)

	//brokerAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeOperations{})
	listener, _ := net.Listen("tcp", *pAddr)
	fmt.Println(*pAddr)
	defer listener.Close()
	rpc.Accept(listener)

}
