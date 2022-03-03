package main

import (
	"net"
	"strings"
)

// user
type User struct {
	Name string
	Addr string      //地址
	C    chan string //用于携程之间的通信
	conn net.Conn    //该用户的套接字

	server *Server //该user对应得server
}

// ListenMessage 监听当前user channl 的方法， 一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

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

//user online
func (this *User) Online() {
	this.server.maplock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.maplock.Unlock()

	this.server.BroadCast(this, "online")
}

//user offine
func (this *User) offine() {
	//用户下线，将用户从onlinemap中删除
	this.server.maplock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.maplock.Unlock()

	this.server.BroadCast(this, "offine")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

//what user to do
func (this *User) DoMessage(msg string) {
	//查询当前在线用户
	if msg == "who" {
		this.server.maplock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMessage := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMessage)
		}
		this.server.maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//此处，split是用在分割数组，以|来分割
		//例如，rename|gzh，分割成 rename， gzh
		newname := strings.Split(msg, "|")[1]
		//判断name是否存在
		//只关心能不能找到
		_, ok := this.server.OnlineMap[newname]
		if ok {
			this.SendMsg("this name has used\n")
		} else {
			this.server.maplock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newname] = this
			this.server.maplock.Unlock()

			this.Name = newname
			this.SendMsg("has rename:" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//格式 to|张三|消息内容
		remoteName := strings.Split(msg, "|")[1]
		//获取要发送的人
		if remoteName == "" {
			this.SendMsg("please enter enough message")
			return
		}
		//根据用户名获取user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("this user didn't exit")
			return
		}
		//获取消息内容，对user发送信息
		remoteMessage := strings.Split(msg, "|")[2]
		remoteUser.SendMsg(this.Name + "say to you: " + remoteMessage)
	} else {
		this.server.BroadCast(this, msg)
	}

}
