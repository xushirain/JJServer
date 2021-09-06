package main

import (
	"fmt"
	"net"
)

type Server struct {
	IP   string
	Port int
}

//创建一个Server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:   ip,
		Port: port,
	}
	return server
}

func (server *Server) Handler(conn net.Conn) {
	//当前连接业务
	fmt.Println("链接建立成功")
}

//启动服务器的端口
func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()
	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do Handler
		go server.Handler(conn)
	}

}
