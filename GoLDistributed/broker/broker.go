package main

import (
	"GameOfLifeReal/stubs"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

var updatedSegments = make([][]byte, 0)
var ln net.Listener
var turn int

type GameOfLifeOperations struct{}

// func makeSegmentByteArray(p stubs.Params /*start and end*/) [][]byte {
// 	newArray := make([][]byte, p.ImageWidth)
// 	for i := 0; i < p.ImageWidth; i++ {
// 		newArray[i] = make([]byte, p.ImageHeight/p.Threads)
// 	}
// 	return newArray
// }
// func that makes a call to the Server; send segment and receive segment

var ip string

func callServer(world [][]byte, p stubs.Params) [][]byte {
	Servers := make([]string, p.Threads)
	for i := 0; i < p.Threads; i++ {
		if i == 1 {
			ip = "3.80.92.80"
		} else if i == 2 {
			ip = "54.152.25.156"
		} else if i == 3 {
			ip = "54.234.171.12"
		}
		server := ip + strconv.Itoa(8031+i)
		flag.Parse()
		fmt.Println("Server: ", server)
		Servers[i] = server
	}

	flag.Parse()
	fmt.Println("Server: ", Servers[0])
	//client, _ := rpc.Dial("tcp", server)

	turn = 0
	//byte array for empty segment
	// segment := makeSegmentByteArray(p)
	segmentHeight := p.ImageHeight / p.Threads

	//response := new(stubs.Response)

	clients := make([]*rpc.Client, p.Threads)
	for i := 0; i < p.Threads; i++ {
		clients[i], _ = rpc.Dial("tcp", Servers[i])

	}

	for turn < p.Turns {
		calls := make([]*rpc.Call, p.Threads)
		responses := make([]*stubs.Response, p.Threads)
		for i := 0; i < p.Threads; i++ {
			responses[i] = new(stubs.Response)
		}

		for i, client := range clients {
			if i == p.Threads-1 {
				request := stubs.Request{World: world, SegStart: segmentHeight * i, SegEnd: p.ImageHeight, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
				//fmt.Println("before client.go")
				calls[i] = client.Go(stubs.GolHandler, request, responses[i], nil)
				fmt.Println("after call")
			} else {
				request := stubs.Request{World: world, SegStart: segmentHeight * i, SegEnd: segmentHeight * (i + 1), P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
				//fmt.Println("before client.go")
				calls[i] = client.Go(stubs.GolHandler, request, responses[i], nil)
				fmt.Println("after call")
			}
		}
		var newWorld [][]byte
		for i, call := range calls {
			<-call.Done
			//fmt.Println("SEGMENT ", i, "  ", responses[i].NewSegment)
			newWorld = append(newWorld, responses[i].NewSegment...)
			//world = newWorld
		}
		world = newWorld
		//fmt.Println(len(world))
		turn++
	}

	// for turn < p.Turns {
	// 	for i := 0; i < p.Threads; i++ {
	// 		if i == p.Threads-1 {
	// 			fmt.Println(Servers[i])
	// 			client, _ := rpc.Dial("tcp", Servers[i])
	// 			fmt.Println("after dial")
	// 			//getting the segment to send
	// 			request := stubs.Request{World: world, Segment: segment, SegStart: segmentHeight * i, SegEnd: p.ImageHeight, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
	// 			//fmt.Println("before client.go")
	// 			call := client.Go(stubs.GolHandler, request, response, nil)
	// 			fmt.Println("after call")
	// 			//fmt.Println("after client.go")
	// 			select {
	// 			case <-call.Done:
	// 				//fmt.Println(response.NewSegment)
	// 				newWorld = append(newWorld, response.NewSegment...)
	// 				world = newWorld
	// 				turn++
	// 			}
	// 		} else {
	// 			fmt.Println(Servers[i])
	// 			client, _ := rpc.Dial("tcp", Servers[i])
	// 			fmt.Println("after dial")
	// 			request := stubs.Request{World: world, Segment: segment, SegStart: segmentHeight * i, SegEnd: segmentHeight*i + 1, P: stubs.Params{ImageHeight: p.ImageHeight, ImageWidth: p.ImageWidth, Threads: p.Threads, Turns: p.Turns}}
	// 			//fmt.Println("before client.go")
	// 			call := client.Go(stubs.GolHandler, request, response, nil)
	// 			fmt.Println("after call")
	// 			//fmt.Println("after client.go")
	// 			select {
	// 			case <-call.Done:
	// 				//fmt.Println(response.NewSegment)
	// 				newWorld = append(newWorld, response.NewSegment...)
	// 				world = newWorld
	// 				turn++
	// 			}
	// 		}
	// 		// defer client.Close()
	// 	}
	// }
	//fmt.Println(len(world))
	return world
}

func (s *GameOfLifeOperations) BrokerProcessGol(req stubs.Request, res *stubs.Response) (err error) {
	//call the split world func
	turn := 0
	fmt.Println("inside exported brokerprocess before server call")
	//call func that sends and receives segment

	newWorld := callServer(req.World, req.P)
	fmt.Println("after callserver")
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
	listener, err := net.Listen("tcp", "127.0.0.1:8030") //"127.0.0.1:"+*pAddr)
	fmt.Println("127.0.0.1:" + *pAddr)
	fmt.Println(err)
	ln = listener
	defer listener.Close()
	rpc.Accept(listener)
}
