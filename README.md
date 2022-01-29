# ServerStatus-goclient

使用Golang写的ServerStatus-Hotaru客户端。

请直接下载[release](https://github.com/cokemine/ServerStatus-goclient/releases)下的对应平台的二进制文件。

## 使用说明

运行时需传入客户端对应参数。

假设你的服务端地址是`yourip`，客户端用户名`username`，密码`password`

端口号`35601`

你可以这样运行

```bash
chmod +x status-client
./status-client -dsn="username:password@yourip:35601"
```

即用户名密码以`:`分割，登录信息和服务器信息以`@`分割，地址与端口号以`:`分割。

默认端口号是35601，所以你可以忽略端口号不写，即直接写`username:password@yourip`

或者使用一键脚本

```shell

wget https://raw.githubusercontent.com/cokemine/ServerStatus-goclient/master/install.sh

#安装

bash install.sh

#或

bash install.sh -dsn username:password@yourip:35601

#修改配置

bash install.sh reset_conf #(re)

#卸载

bash install.sh uninstall #(uni)

```

## Usage

```
  -dsn string
        Input DSN, format: username:password@host:port
  -h string
        Input the host of the server
  -interval float
        Input the INTERVAL (default 2.0)
  -p string
        Input the client's password
  -port int
        Input the port of the server (default 35601)
  -u string
        Input the client's username
  -vnstat
        Use vnstat for traffic statistics, linux only
```

