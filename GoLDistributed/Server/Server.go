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

func calculateNextState(p stubs.Params, world [][]byte) [][]byte {
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

				} else if sum == 2 || sum == 3 {
					newWorld[y][x] = 255
				} else {
					newWorld[y][x] = 0

				}
			} else {
				if sum == 3 {
					newWorld[y][x] = 255

				} else {
					newWorld[y][x] = 0
				}
			}
		}
	}
	return newWorld
}

type GameOfLifeOperations struct{}

func (s *GameOfLifeOperations) ProcessGameOfLife(req stubs.Request, res *stubs.Response) (err error) {

	newWorld := req.World

	if req.P.Turns != 0 {
		newWorld = calculateNextState(req.P, newWorld)
	}
	res.NewWorld = newWorld
	return
}

func (s *GameOfLifeOperations) KillProcess(req stubs.Request, res stubs.Response) (err error) {
	ln.Close()
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	ln = listener
	defer listener.Close()
	rpc.Accept(listener)

}
