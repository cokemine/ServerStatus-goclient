package status

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	psutilNet "github.com/shirou/gopsutil/v3/net"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var cachedFs = make(map[string]struct{})
var timer = 0.0

type network struct {
	rx *deque
	tx *deque
}

func NewNetwork() *network {
	instance := &network{
		newDeque(10),
		newDeque(10),
	}
	return instance
}

func Uptime() uint64 {
	uptime, _ := host.Uptime()
	return uptime
}

func Memory() (uint64, uint64, uint64, uint64) {
	memory, _ := mem.VirtualMemory()
	swap, _ := mem.SwapMemory()
	return memory.Total / 1024.0, memory.Used / 1024.0, swap.Total / 1024.0, swap.Used / 1024.0
}

func Load() float64 {
	theLoad, _ := load.Avg()
	return theLoad.Load1
}

func Disk(INTERVAL *float64) (uint64, uint64) {
	var (
		size, used uint64
	)
	if timer <= 0 {
		diskList, _ := disk.Partitions(false)
		for _, d := range diskList {
			if checkValidFs(d.Fstype) {
				cachedFs[d.Mountpoint] = struct{}{}
			}
		}
		timer = 150.0
	}
	timer -= *INTERVAL
	for k := range cachedFs {
		usage, _ := disk.Usage(k)
		size += usage.Total / 1024.0 / 1024.0
		used += usage.Used / 1024.0 / 1024.0
	}
	return size, used
}

func Cpu(INTERVAL *float64) float64 {
	cpuInfo, _ := cpu.Percent(time.Duration(*INTERVAL)*time.Second, true)
	return cpuInfo[0]
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
	err = conn.Close()
	if err != nil {
		return false
	}
	return true
}

func (net *network) getTraffic() {
	var (
		netIn, netOut uint64
	)
	netInfo, _ := psutilNet.IOCounters(true)
	for _, v := range netInfo {
		if checkInterface(v.Name) {
			netIn += v.BytesRecv
			netOut += v.BytesSent
		}
	}
	net.rx.push(netIn)
	net.tx.push(netOut)
}

func (net *network) Traffic() (uint64, uint64) {
	return net.rx.tail.value, net.tx.tail.value
}

func TrafficVnstat() (uint64, uint64, error) {
	buf, err := exec.Command("vnstat", "--oneline", "b").Output()
	if err != nil {
		return 0, 0, err
	}
	vData := strings.Split(BytesToString(buf), ";")
	rx, err := strconv.ParseUint(vData[8], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	tx, err := strconv.ParseUint(vData[9], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return rx, tx, nil
}

func (net *network) Speed() (uint64, uint64) {
	net.getTraffic()
	return uint64(net.rx.avg()), uint64(net.tx.avg())
}
