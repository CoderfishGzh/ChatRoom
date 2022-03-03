package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	Name       string
	Conn       net.Conn
	ServerIp   string
	ServerPort int
	clientflag int
}

var serverIP string
var serverPort int

func NewClient(ip string, port int) *Client {
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		clientflag: 999,
	}
	//向指定地址：端口发起tcp连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	//如果err不为nil的话，就代表连接发生错误
	if err != nil {
		fmt.Println("net Dial error: ", err)
		return nil
	}

	client.Conn = conn

	return client
}

func (this *Client) menu() bool {
	var clientflag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.改用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&clientflag) //从键盘中读取参数到clientfalg中，且一直读取到\n出，即每次换行读取一次
	if err != nil {
		fmt.Println("scanln err ", err)
		return false
	}
	fmt.Println("test----", clientflag)
	fmt.Println("test---参数", clientflag)
	if clientflag >= 0 && clientflag <= 3 {
		this.clientflag = clientflag
		return true
	} else {
		fmt.Println("请输入范围内的flag")
		return false
	}
}

//更新用户名
func (this *Client) UpdateName() bool {
	//封装通信格式
	fmt.Println("请输入要更改的名字...")
	fmt.Scanln(&this.Name)

	msg := "rename|" + this.Name + "\n"
	_, err := this.Conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn write error", err)
		return false
	}
	return true
}

//公聊模式
func (this *Client) PublicChat() {
	fmt.Println("public chat--enter“exit” exit")
	var entermsg string
	fmt.Scanln(&entermsg)

	for entermsg != "exit" {
		//消息不为空则发送给给服务器
		if len(entermsg) != 0 {
			msg := entermsg + "\n"
			_, err := this.Conn.Write([]byte(msg))
			if err != nil {
				fmt.Println("conn write error...")
				break
			}
		}

		entermsg = ""
		fmt.Println("public chat--enter“exit” exit")
		fmt.Scanln(&entermsg)
	}
}

//查询在线用户信息
func (this *Client) WhoOnline() {
	msg := "who\n"
	_, err := this.Conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn write err ", err)
		return
	}
}

//私人聊天模式
func (this *Client) PrivteChat() {
	//私聊之前应该获取当前在线的用户
	var RemoteUser string
	var msg string
	this.WhoOnline()
	fmt.Println("请选择要私聊的用户，exit退出")
	fmt.Scanln(&RemoteUser)

	//2 循环获取想要私聊的用户
	for RemoteUser != "exit" {
		//3 循环获取要发送的信息
		fmt.Println("请输入要发送的内容, exit退出")
		fmt.Scanln(&msg)

		for msg != "exit" {
			remotemsg := "to|" + RemoteUser + "|" + msg + "\n\n"
			_, err := this.Conn.Write([]byte(remotemsg))
			if err != nil {
				fmt.Println("Conn write err ", err)
				break
			}

			remotemsg = ""
			fmt.Println("请输入要发送的内容, exit退出")
			fmt.Scanln(&msg)
		}
		RemoteUser = ""
		fmt.Println("请选择要私聊的用户，exit退出")
		fmt.Scanln(&RemoteUser)
	}
}

//用于客户端业务
func (this *Client) run() {
	//循环判断，如果flag == 0 即是退出
	for this.clientflag != 0 {
		//一直进行循环，直到menu返回true，即知道用户输入正确的名命令才进行操作
		for this.menu() == false {

		}
		switch this.clientflag {
		case 1:
			fmt.Println("public chat")
			this.PublicChat()
			break
		case 2:
			fmt.Println("privte chat")
			this.PrivteChat()
			break
		case 3:
			fmt.Println("rename")
			this.UpdateName()
			break
		}
	}
}

func (this *Client) ServerMessge() {
	//从套接字读出来到标准输出stdout 即屏幕
	io.Copy(os.Stdout, this.Conn)
}

//init 函数是执行在main函数之前
func init() {
	//使用flag库来接收且命令行参数
	flag.StringVar(&serverIP, "IP", "127.0.0.1", "设置服务器IP地址，默认是127.0.0.1")
	flag.IntVar(&serverPort, "Port", 8000, "设置服务器端口，默认是8000")
}
func main() {
	//解析命令行函数
	flag.Parse()

	client := NewClient(serverIP, serverPort)

	if client == nil {
		fmt.Println("链接失败。。。。。")
		return
	}
	fmt.Println("链接成功。。。。")
	go client.ServerMessge()
	//先阻塞
	//select {}

	client.run()
}
