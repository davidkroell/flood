//+build linux

//Copyright (C) 2019  David Kröll
//
//This program is free software: you can redistribute it and/or modify
//it under the terms of the GNU General Public License as published by
//the Free Software Foundation, either version 3 of the License, or
//(at your option) any later version.
//
//This program is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//GNU General Public License for more details.

package main

import (
	"flag"
	"fmt"
	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/raw"
	"log"
	"net"
	"sync"
	"time"
)

var num = flag.Int("n", 1, "Amount of ethernet frames send")
var ifaceIndex = flag.Int("i", 1, "Interface to send")
var listIfaces = flag.Bool("l", false, "List available interfaces")
var numThreads = flag.Int("t", 12, "Number of threads to use")

const etherType = 0xbeef

func main() {
	flag.Parse()

	if *listIfaces {
		ifaces, err := net.Interfaces()
		if err != nil {

		}

		fmt.Println("Index\tName\tAddress")
		format := "%d\t%s\t%s\n"

		for _, iface := range ifaces {
			fmt.Printf(format, iface.Index, iface.Name, iface.HardwareAddr)
		}
		return
	}

	iface, err := net.InterfaceByIndex(*ifaceIndex)
	if err != nil {
		log.Fatalf("Cannot get interface: %v", err)
	}

	conn, err := raw.ListenPacket(iface, etherType, nil)
	if err != nil {
		log.Fatalf("cannot open connection: %v", err)
	}

	var wg sync.WaitGroup

	// init channels
	ch := make(chan *ethernet.Frame)
	closer := make(chan struct{})
	stats := make(chan int)

	// create sender goroutines
	for i := 0; i < *numThreads; i++ {
		go frameWriter(conn, ch, closer, stats)
	}

	// stat collecting goroutine
	go func() {
		wg.Add(1)
		// init stat vars
		framesSend := 0
		bytesWritten := 0
		startTime := time.Now()

		for {
			select {
			case <-closer:
				fmt.Println("---------- SUMMARY ----------")
				fmt.Printf("%d frames send\n", framesSend)
				fmt.Printf("%d bytes written\n", bytesWritten)
				fmt.Printf("Took a total time of %v\n", time.Since(startTime))

				// done
				wg.Done()
				return
			case bytes := <-stats:
				bytesWritten += bytes
				framesSend++
			}
		}
	}()

	for i := 1; i <= *num; i++ {
		f := &ethernet.Frame{
			Destination: ethernet.Broadcast,
			Source:      net.HardwareAddr{0xbe, 0xef, byte(i / 16777216), uint8(i / 65536), uint8(i / 256), uint8(i)},
			EtherType:   etherType,
		}
		ch <- f
	}

	close(closer)
	wg.Wait() // wait for goroutines quit
}

func frameWriter(c net.PacketConn, ch <-chan *ethernet.Frame, closer <-chan struct{}, stats chan<- int) {
	for {
		select {
		case f := <-ch:
			// get frame and marshall it to binary
			b, err := f.MarshalBinary()
			if err != nil {
				log.Fatalf("failed to marshal ethernet frame: %v", err)
			}

			addr := &raw.Addr{
				HardwareAddr: ethernet.Broadcast,
			}

			// write frame
			n, err := c.WriteTo(b, addr)
			if err != nil {
				log.Fatalf("Cannot write to connection: %v", err)
			}

			// send to channel
			stats <- n
		case <-closer:
			return
		}
	}
}
