package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

var ConnectCounter int = 0
var ReceiveCounter int = 0
var SendCounter int = 0
var sn sync.RWMutex

func run(c *cli.Context) error {
	port := c.Int("port")
	n := c.Int("number")
	for i := 0; i < n; i++ {
		go startServer(port + i)
	}
	//quit when receive end signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	fmt.Printf("signal received signal %v\n", <-sigChan)
	fmt.Println("shutting down server")
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "c1000k-server"
	app.Usage = "c1000k-server"
	app.Copyright = "panyingyun@gmail.com"
	app.Version = "0.1.0"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "port,p",
			Usage:  "Set Server start port here",
			Value:  9999,
			EnvVar: "PORT",
		},
		cli.IntFlag{
			Name:   "number,n",
			Usage:  "Set Number of Server here",
			Value:  1,
			EnvVar: "NUMBER",
		},
	}
	app.Run(os.Args)
}

//启动TCP服务
func startServer(port int) {
	addr := fmt.Sprintf("0.0.0.0:%v", port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		handleServerError(err)
		return
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		handleServerError(err)
		return
	}

	defer tcpListener.Close()

	fmt.Printf("Start Server Listen On Address:[%v]\n", addr)
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
		sn.Lock()
		ConnectCounter++
		fmt.Printf("Connect [%v] [%v] [%v] [%v]\n", ConnectCounter, ReceiveCounter, SendCounter, tcpConn.RemoteAddr().String())
		sn.Unlock()

		go handleMessage(tcpConn)
	}
}

func handleMessage(conn *net.TCPConn) {
	ipStr := conn.RemoteAddr().String()
	defer func() {
		sn.Lock()
		ConnectCounter--
		fmt.Printf("Connect [%v] [%v] [%v] [%v]\n", ConnectCounter, ReceiveCounter, SendCounter, ipStr)
		conn.Close()
		sn.Unlock()
	}()
	reader := bufio.NewReader(conn)

	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		sn.Lock()
		ReceiveCounter++
		sn.Unlock()
		//fmt.Println(message)
		msg := time.Now().String() + "\n"
		b := []byte(msg)
		conn.Write(b)
		sn.Lock()
		SendCounter++
		sn.Unlock()
		fmt.Printf("Connect [%v] [%v] [%v] [%v]\n", ConnectCounter, ReceiveCounter, SendCounter, ipStr)
	}
}

func handleServerError(err error) {
	fmt.Println("Server Error:", err)
	fmt.Printf("Connect [%v] [%v] [%v]\n", ConnectCounter, ReceiveCounter, SendCounter)
}