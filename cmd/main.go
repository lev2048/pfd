package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	pfd "pfd/pkg"
	"syscall"
)

var (
	localAddr  string
	remoteAddr string
)

func init() {
	flag.StringVar(&localAddr, "l", "", "local address eg: 0.0.0.0:808")
	flag.StringVar(&remoteAddr, "r", "", "remote address eg: 1.1.1.1:80 or baidu.com:80")
}
func main() {
	flag.Parse()
	if pd, err := pfd.NewPFD(localAddr, remoteAddr); err != nil {
		fmt.Println("init err: ", err)
		os.Exit(-1)
	} else {
		pd.Start()
		fmt.Println("start forward: ", localAddr+" => "+remoteAddr)
		defer pd.Stop()
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
