package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"http-client-example/stoppableListener"
)

var ServerHost string = "localhost"
var ServerPort int = 8008

func ImmediateReturnHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf(">>>>>> This is a immediate return test!\n")
	w.Write([]byte("This is a immediate return test!"))
}

func NeverReturnHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf(">>>>>> This is a never return test!\n")
	s := 0
	for {
		time.Sleep(1 * time.Second)
		s += 1
		fmt.Printf(">>>>>> Never return: sleep %d seconds\n", s)
	}
	w.Write([]byte("This is a never return test!"))
}

func FiveSecondsReturnHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf(">>>>>> This is a 5 seconds return test!\n")
	time.Sleep(5 * time.Second)
	w.Write([]byte("This is a 5 seconds return test!"))
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	originalListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ServerHost, ServerPort))
	if err != nil {
		panic(err)
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/immediate_return", ImmediateReturnHandler)
	http.HandleFunc("/never_return", NeverReturnHandler)
	http.HandleFunc("/5_seconds_return", FiveSecondsReturnHandler)

	server := http.Server{}

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		server.Serve(sl)
	}()

	fmt.Printf("Serving HTTP\n")

	select {
	case signal := <-sigs:
		fmt.Printf("Got signal: (%v)\n", signal)
	}

	time.Sleep(5)
	sl.Stop()

	wg.Wait()
	fmt.Printf("Server exit!\n")
	os.Exit(0)
}
