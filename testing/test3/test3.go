package main

import (
	"fmt"
	"time"
)

type Server struct {
	Name    string
	Message chan string
	Signal  chan bool
}

func main() {
	var serverList []*Server

	s1 := &Server{
		Name:    "s1",
		Message: make(chan string),
		Signal:  make(chan bool),
	}
	serverList = append(serverList, s1)

	s2 := &Server{
		Name:    "s2",
		Message: make(chan string),
		Signal:  make(chan bool),
	}
	serverList = append(serverList, s2)

	s3 := &Server{
		Name:    "s3",
		Message: make(chan string),
		Signal:  make(chan bool),
	}
	serverList = append(serverList, s3)

	for _, s := range serverList {
		go server(s)
	}

	s1.Message <- s1.Name

	time.Sleep(time.Duration(10) * time.Second)

	s1.Signal <- true

	time.Sleep(time.Duration(10) * time.Second)

	s1.Message <- s1.Name

}

func server(s *Server) {
	for {
		select {
		case msg := <-s.Message:
			fmt.Println("received message", msg)
		case sig := <-s.Signal:
			fmt.Println("received signal", sig)
			close(s.Message)
			return
		default:
		}

	}
}
