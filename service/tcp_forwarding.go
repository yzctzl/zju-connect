package service

import (
	"github.com/mythologyli/zju-connect/log"
	"github.com/mythologyli/zju-connect/stack"
	"io"
	"net"
	"strconv"
	"strings"
)

func handleRequest(stack stack.Stack, conn net.Conn, remoteAddress string) {
	log.Printf("Port forwarding (TCP): %s -> %s -> %s", conn.RemoteAddr(), conn.LocalAddr(), remoteAddress)

	parts := strings.Split(remoteAddress, ":")
	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Port forwarding (TCP) invalid port: %v", err)
		_ = conn.Close()
		return
	}

	proxy, err := stack.DialTCP(&net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	})
	if err != nil {
		log.Printf("Port forwarding (TCP) failed to dial %s: %v", remoteAddress, err)
		_ = conn.Close()
		return
	}

	go copyIO(conn, proxy)
	go copyIO(proxy, conn)
}

func copyIO(src, dest net.Conn) {
	defer func(src net.Conn) {
		_ = src.Close()
	}(src)
	defer func(dest net.Conn) {
		_ = dest.Close()
	}(dest)
	_, _ = io.Copy(src, dest)
}

func ServeTCPForwarding(stack stack.Stack, bindAddress string, remoteAddress string) {
	ln, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatalf("TCP port forwarding listen failed: %v", err)
	}

	log.Printf("TCP port forwarding: %s -> %s", bindAddress, remoteAddress)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("TCP port forwarding accept error: %v", err)
			continue
		}

		go handleRequest(stack, conn, remoteAddress)
	}
}
