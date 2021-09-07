package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个Server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听Message
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		//将msg发送给全部在线的User
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

//广播消息的方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//当前连接业务
	//fmt.Println("链接建立成功")
	user := NewUser(conn, server)

	user.Online()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	//接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的消息(去除\n)
			msg := string(buf[:n-1])

			//用户针对msg进行消息处理
			user.DoMessage(msg)

			//用户的任意消息代表当前用户是活跃的
			isLive <- true
		}
	}()

	//当前Handler阻塞
	for {
		select {
		case <-isLive:
			//当前用户是活跃的，应当重置定时器
			//不做任何事，为了激活select，更新下边的定时器
		case <-time.After(time.Second * 300):
			//已经超时
			//将当前的User强制的关闭
			user.SendMsg("你被踢了")

			//销毁用的资源
			close(user.C)

			//关闭连接
			conn.Close()

			//退出当前handler
			return //runtime.Goexit()
		}
	}
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

	go server.ListenMessage()

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
