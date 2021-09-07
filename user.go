package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

//用户上线
func (user *User) Online() {
	//用户上线
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	//广播当前用户上线消息
	user.server.BroadCast(user, "已上线")
}

//用户下线
func (user *User) Offline() {
	//用户下线
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	//广播当前用户上线消息
	user.server.BroadCast(user, "下线")
}

//给当前User对应的客户端发送消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

//用户处理消息
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前有哪些用户
		user.server.mapLock.Lock()
		for _, user := range user.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[0:7] == "rename" {
		newName := strings.Split(msg, "|")[1]
		//判断name是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("当前用户名被使用\n")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.Name = newName
			user.SendMsg("您已经更新用户名:" + user.Name + "\n")
		}
	} else if len(msg) > 4 && msg[0:3] == "to" {
		//1 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("消息格式不正确，请使用 \"to|张三|你好\"格式,\n")
		}
		//2 根据用户名 得到对方User对象
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("该用户名不存在\n")
			return
		}
		//3 获取消息内容 通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("无消息内容,请重发\n")
			return
		}
		remoteUser.SendMsg(user.Name + "对您说:" + content)
	} else {
		user.server.BroadCast(user, msg)
	}
}

//监听当前User channel的方法，一旦有消息就直接发送给对端客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
