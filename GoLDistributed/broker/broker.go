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

var updatedSegments = make([][]byte, 0)
var ln net.Listener
var turn int

type GameOfLifeOperations struct{}

func makeSegmentByteArray(p stubs.Params /*start and end*/) [][]byte {
	newArray := make([][]byte, p.ImageWidth)
	for i := 0; i < p.ImageWidth; i++ {
		newArray[i] = make([]byte, p.ImageHeight/p.Threads)
	}
	return newArray
}

// func that makes a call to the Server; send segment and receive segment
func callServer(world [][]byte, p stubs.Params) [][]byte {
	server := "127.0.0.1:8031"
	flag.Parse()
	fmt.Println("Server: ", server)
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()
	var newWorld [][]byte
	turn = 0
	//byte array for empty segment
	segment := makeSegmentByteArray(p)
	segmentHeight := p.ImageHeight / p.Threads

	response := new(stubs.Response)
	for turn < p.Turns {
		for i := 0; i < p.Threads; i++ {
			if i == p.Threads-1 {
				//getting the segment to send
				request := stubs.Request{World: world, Segment: segment, SegStart: segmentHeight * i, SegEnd: p.ImageHeight, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
				//fmt.Println("before client.go")
				call := client.Go(stubs.GolHandler, request, response, nil)
				//fmt.Println("after client.go")
				select {
				case <-call.Done:
					//fmt.Println(response.NewSegment)
					newWorld = append(newWorld, response.NewSegment...)
					world = newWorld
					turn++
				}
			} else {
				request := stubs.Request{World: world, Segment: segment, SegStart: segmentHeight * i, SegEnd: segmentHeight*i + 1, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
				//fmt.Println("before client.go")
				call := client.Go(stubs.GolHandler, request, response, nil)
				//fmt.Println("after client.go")
				select {
				case <-call.Done:
					//fmt.Println(response.NewSegment)
					newWorld = append(newWorld, response.NewSegment...)
					world = newWorld
					turn++
				}
			}
		}
	}
	return world
}

func (s *GameOfLifeOperations) BrokerProcessGol(req stubs.Request, res *stubs.Response) (err error) {
	//call the split world func
	turn := 0
	fmt.Println("inside exported brokerprocess before server call")
	//call func that sends and receives segment

	newWorld := callServer(req.World, req.P)
	//put segments back togther and send back updated world
	res.NewWorld = newWorld
	turn++

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
