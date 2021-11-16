package pfd

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type PFD struct {
	ldr     string
	dst     string
	domain  string
	exitTcp chan struct{}
}

func NewPFD(ldr, dst string) (*PFD, error) {
	if dest, dmn, err := resolveIp(dst); err != nil {
		return nil, err
	} else {
		return &PFD{ldr, dest, dmn, make(chan struct{})}, nil
	}
}

func resolveIp(addr string) (string, string, error) {
	str := strings.Split(addr, ":")
	if len(str) != 2 {
		return "", "", errors.New("addr illegal " + addr)
	}
	host, port := str[0], str[1]
	if ip := net.ParseIP(host); ip != nil {
		return addr, "", nil
	} else {
		if ips, err := net.LookupIP(host); err != nil {
			return "", "", err
		} else {
			return ips[0].String() + ":" + port, addr, nil
		}
	}
}

func (p *PFD) Start() {
	if p.domain != "" {
		go p.checkDomain()
	}
	go p.serveTcp()
}

func (p *PFD) serveTcp() {
	ln, err := net.Listen("tcp", p.ldr)
	if err != nil {
		fmt.Println("listenTcp err:", err)
		return
	}
	defer ln.Close()
	for {
		select {
		case <-p.exitTcp:
			fmt.Println("exitTcp")
			return
		default:
			if conn, err := ln.Accept(); err == nil {
				go p.handleTcpConn(conn)
			} else {
				fmt.Println("accept err:", err)
			}
		}
	}
}

func (p *PFD) checkDomain() {
	for {
		time.Sleep(5 * time.Minute)
		select {
		case <-p.exitTcp:
			return
		default:
			if ips, err := net.LookupIP(strings.Split(p.domain, ":")[0]); err != nil {
				fmt.Println("check domain err: ", err)
			} else {
				if ips[0].String() != p.dst {
					p.dst = ips[0].String() + ":" + strings.Split(p.domain, ":")[1]
				}
			}
		}
	}
}

func (p *PFD) handleTcpConn(conn net.Conn) {
	rcon, err := net.Dial("tcp", p.dst)
	if err != nil {
		fmt.Println("dial remote error: ", err)
		return
	}
	wg, isClose := &sync.WaitGroup{}, false
	sendBuf, recvBuf := make([]byte, 1460), make([]byte, 1460)
	conn.SetReadDeadline(time.Now().Add(time.Second * 15))
	rcon.SetReadDeadline(time.Now().Add(time.Second * 15))
	wg.Add(2)
	go func() {
		defer func() {
			isClose = true
			conn.Close()
			rcon.Close()
			wg.Done()
		}()
		for {
			if n, err := conn.Read(sendBuf); err != nil {
				if isClose || err == io.EOF {
					return
				}
				fmt.Println("read local tcp err: ", err, isClose)
				return
			} else {
				conn.SetReadDeadline(time.Now().Add(time.Second * 15))
				if _, err := rcon.Write(sendBuf[:n]); err != nil {
					if isClose || err == io.EOF {
						return
					}
					fmt.Println("send remote tcp err: ", err)
					return
				}
			}
		}
	}()
	go func() {
		defer func() {
			isClose = true
			conn.Close()
			rcon.Close()
			wg.Done()
		}()
		for {
			if n, err := rcon.Read(recvBuf); err != nil {
				if isClose || err == io.EOF {
					return
				}
				fmt.Println("read remote tcp err: ", err, isClose)
				return
			} else {
				rcon.SetReadDeadline(time.Now().Add(time.Second * 15))
				if _, err := conn.Write(recvBuf[:n]); err != nil {
					if isClose || err == io.EOF {
						return
					}
					fmt.Println("send local tcp err: ", err)
					return
				}
			}
		}
	}()
	wg.Wait()
}

func (p *PFD) Stop() error {
	close(p.exitTcp)
	return nil
}
