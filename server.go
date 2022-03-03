package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

//服务器server
type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	maplock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//server的创建
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg

}

func (this *Server) Hander(conn net.Conn) {
	//fmt.Println("链接建立成功")
	//链接成功的业务
	user := NewUser(conn, this)

	user.Online()

	islive := make(chan bool)

	//一直循环读取
	go func() {
		buf := make([]byte, 4069)
		for {
			//n是读的字节数， err是读期间发生的错误
			n, err := conn.Read(buf)
			if n == 0 {
				user.offine()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err: ", err)
				return
			}

			msg := string(buf[:n-1])
			user.DoMessage(msg)
			//发送信息后，把true推到islive的channl中
			islive <- true
			fmt.Println(" server test...client msg is ", msg)
		}
	}()

	//select {}
	//这里使用select来堵塞，是因为，如果此hander解释后
	//gc会回收该 user的资源,后续调用 *user时，会发生错误
	//当前handle阻塞
	for {
		select {
		case <-islive:

		case <-time.After(time.Second * 300):
			//已经超时
			user.SendMsg("you has out")
			close(user.C)
			//关闭conn后，读取conn时，会返回0，跟胡handle逻辑，会调用offine函数进行用户的下线
			conn.Close()
			return
		}
	}
}

func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.maplock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.maplock.Unlock()
	}
}

// Start 启动服务器
func (this Server) Start() {
	//监听服务器的 ip和端口
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen error..")
		fmt.Println(err)
		return
	}
	defer listener.Close()

	go this.ListenMessager()
	//监听成功
	for {
		//accept 接受
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error...")
			fmt.Println("err is : ", err)
			return
		}
		//处理此连接，既处理客户端发送过来的链接
		go this.Hander(conn)
	}

}
