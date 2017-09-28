package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const GOSH_ENV = "SS_SERVER"

func main() {
	// start ss http proxy
	port := startSSClient()
	// copy cmd to go
	cmd := exec.Command("go", os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	wd, _ := os.Getwd()
	cmd.Dir = wd
	cmd.Env = append(os.Environ(), "HTTP_PROXY=http://127.0.0.1:"+strconv.Itoa(port), "HTTPS_PROXY=http://127.0.0.1:"+strconv.Itoa(port))
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
func startSSClient() (port int) {

	ssServer := os.Getenv(GOSH_ENV)

	if ssServer == "" {
		fmt.Println("MUST HAS " + GOSH_ENV + " in env")
		os.Exit(-1)
	}
	u, err := url.Parse(ssServer)
	if err != nil {
		fmt.Println(GOSH_ENV+" wrong", err)
		os.Exit(-1)
	}
	password, _ := u.User.Password()
	cipher, err := shadowsocks.NewCipher(u.User.Username(), password)
	if err != nil {
		fmt.Println(GOSH_ENV+" wrong", err)
		os.Exit(-1)
	}
	dialer, err := shadowsocks.NewDialer(u.Host, cipher)
	if err != nil {
		fmt.Println(GOSH_ENV+" wrong", err)
		os.Exit(-1)
	}
	server := goproxy.NewProxyHttpServer()

	server.ConnectDial = dialer.Dial
	server.Tr.Dial = dialer.Dial
	return listenHTTPProxy(server)
}

func listenHTTPProxy(server http.Handler) (port int) {
	var ln net.Listener
	var err error
	for i := 0; i < 65535-3000; i++ {
		port = 3000 + i
		ln, err = net.Listen("tcp", ":"+strconv.Itoa(port))
		if err == nil {
			break
		}
	}
	if err != nil {
		panic("cannot found a port to listen")
	}
	go http.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}, server)
	return port
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
