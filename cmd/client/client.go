package main

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/cokemine/ServerStatus-goclient/pkg/status"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	SERVER   = flag.String("h", "", "Input the host of the server")
	PORT     = flag.Int("port", 35601, "Input the port of the server")
	USER     = flag.String("u", "", "Input the client's username")
	PASSWORD = flag.String("p", "", "Input the client's password")
	INTERVAL = flag.Float64("interval", 1.5, "Input the INTERVAL")
	DSN      = flag.String("dsn", "", "Input DSN, format: username:password@host:port")
	isVnstat = flag.Bool("vnstat", false, "Use vnstat for traffic statistics, linux only")
)

type serverStatus struct {
	Uptime      uint64  `json:"uptime"`
	Load        float64 `json:"load"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	HddTotal    uint64  `json:"hdd_total"`
	HddUsed     uint64  `json:"hdd_used"`
	CPU         float64 `json:"cpu"`
	NetworkTx   uint64  `json:"network_tx"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkIn   uint64  `json:"network_in"`
	NetworkOut  uint64  `json:"network_out"`
	Online4     bool    `json:"online4"`
	Online6     bool    `json:"online6"`
}

func main() {
	flag.Parse()
	if *DSN != "" {
		dsn := strings.Split(*DSN, "@")
		prev := strings.Split(dsn[0], ":")
		next := strings.Split(dsn[1], ":")
		*USER = prev[0]
		*PASSWORD = prev[1]
		*SERVER = next[0]
		if len(next) == 2 {
			*PORT, _ = strconv.Atoi(next[1])
		}
	}
	if *PORT < 1 || *PORT > 65535 {
		log.Println("Check the port you input")
		os.Exit(1)
	}
	if *SERVER == "" || *USER == "" || *PASSWORD == "" {
		log.Println("HOST, USERNAME, PASSWORD must not be blank!")
		os.Exit(1)
	}
	for {
		log.Println("Connecting...")
		conn, err := net.DialTimeout("tcp", *SERVER+":"+strconv.Itoa(*PORT), 30*time.Second)
		if err != nil {
			log.Println("Caught Exception:", err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		var buf [1024]byte
		var data = status.BytesToString(buf[:])
		n, _ := conn.Read(buf[:])
		if !strings.Contains(data, "Authentication required") || err != nil {
			e := err
			if e == nil {
				log.Println(data[:n])
				e = errors.New("authentication error")
			}
			log.Println("Caught Exception:", e.Error())
			time.Sleep(5 * time.Second)
			continue
		} else {
			_, _ = conn.Write([]byte((*USER + ":" + *PASSWORD + "\n")))
		}
		n, _ = conn.Read(buf[:])
		log.Println(data[:n])
		if !strings.Contains(data, "You are connecting via") {
			n, err = conn.Read(buf[:])
			log.Println(data[:n])
		}
		timer := 0.0
		checkIP := 0
		if strings.Contains(data, "IPv4") {
			checkIP = 6
		} else if strings.Contains(data, "IPv6") {
			checkIP = 4
		} else {
			log.Println(data[:n])
			time.Sleep(5 * time.Second)
			continue
		}
		item := &serverStatus{}
		traffic := status.NewNetwork()
		for {
			CPU := status.Cpu(INTERVAL)
			netRx, netTx := traffic.Speed()
			var netIn, netOut uint64
			if !*isVnstat {
				netIn, netOut = traffic.Traffic()
			} else {
				netIn, netOut, err = status.TrafficVnstat()
				if err != nil {
					log.Println("Please check if the installation of vnStat is correct")
					os.Exit(1)
				}
			}
			memoryTotal, memoryUsed, swapTotal, swapUsed := status.Memory()
			hddTotal, hddUsed := status.Disk(INTERVAL)
			uptime := status.Uptime()
			load := status.Load()
			item.CPU = CPU
			item.Load = load
			item.Uptime = uptime
			item.MemoryTotal = memoryTotal
			item.MemoryUsed = memoryUsed
			item.SwapTotal = swapTotal
			item.SwapUsed = swapUsed
			item.HddTotal = hddTotal
			item.HddUsed = hddUsed
			item.NetworkRx = netRx
			item.NetworkTx = netTx
			item.NetworkIn = netIn
			item.NetworkOut = netOut
			if timer <= 0 {
				if checkIP == 4 {
					item.Online4 = status.Network(checkIP)
					item.Online6 = true
				} else if checkIP == 6 {
					item.Online6 = status.Network(checkIP)
					item.Online4 = true
				}
				timer = 150.0
			}
			timer -= 1 * *INTERVAL
			data, _ := json.Marshal(item)
			_, err = conn.Write(status.StringToBytes("update " + status.BytesToString(data) + "\n"))
			if err != nil {
				log.Println(err.Error())
				break
			}
		}
	}
}
