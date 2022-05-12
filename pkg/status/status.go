package status

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	pNet "github.com/shirou/gopsutil/v3/net"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var cachedFs = make(map[string]struct{})
var timer = 0.0
var prevNetIn uint64
var prevNetOut uint64

func Uptime() uint64 {
	bootTime, _ := host.BootTime()
	return uint64(time.Now().Unix()) - bootTime
}

func Load() float64 {
	theLoad, _ := load.Avg()
	return theLoad.Load1
}

func Disk(INTERVAL float64) (uint64, uint64) {
	var (
		size, used uint64
	)
	if timer <= 0 {
		diskList, _ := disk.Partitions(false)
		devices := make(map[string]struct{})
		for _, d := range diskList {
			_, ok := devices[d.Device]
			if !ok && checkValidFs(d.Fstype) {
				cachedFs[d.Mountpoint] = struct{}{}
				devices[d.Device] = struct{}{}
			}
		}
		timer = 300.0
	}
	timer -= INTERVAL
	for k := range cachedFs {
		usage, err := disk.Usage(k)
		if err != nil {
			delete(cachedFs, k)
			continue
		}
		size += usage.Total / 1024.0 / 1024.0
		used += usage.Used / 1024.0 / 1024.0
	}
	return size, used
}

func Cpu(INTERVAL float64) float64 {
	cpuInfo, _ := cpu.Percent(time.Duration(INTERVAL*float64(time.Second)), false)
	return math.Round(cpuInfo[0]*10) / 10
}

func Network(checkIP int) bool {
	var HOST string
	if checkIP == 4 {
		HOST = "8.8.8.8:53"
	} else if checkIP == 6 {
		HOST = "[2001:4860:4860::8888]:53"
	} else {
		return false
	}
	conn, err := net.DialTimeout("tcp", HOST, 2*time.Second)
	if err != nil {
		return false
	}
	if conn.Close() != nil {
		return false
	}
	return true
}

func Traffic(INTERVAL float64) (uint64, uint64, uint64, uint64) {
	var (
		netIn, netOut uint64
	)
	netInfo, _ := pNet.IOCounters(true)
	for _, v := range netInfo {
		if checkInterface(v.Name) {
			netIn += v.BytesRecv
			netOut += v.BytesSent
		}
	}
	rx := uint64(float64(netIn-prevNetIn) / INTERVAL)
	tx := uint64(float64(netOut-prevNetOut) / INTERVAL)
	prevNetIn = netIn
	prevNetOut = netOut
	return netIn, netOut, rx, tx
}

func TrafficVnstat() (uint64, uint64, error) {
	buf, err := exec.Command("vnstat", "--oneline", "b").Output()
	if err != nil {
		return 0, 0, err
	}
	vData := strings.Split(BytesToString(buf), ";")
	if len(vData) != 15 {
		// Not enough data available yet.
		return 0, 0, nil
	}
	netIn, err := strconv.ParseUint(vData[8], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	netOut, err := strconv.ParseUint(vData[9], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return netIn, netOut, nil
}
