package helper

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func GetCPUUsage() ([]float64, error) {
	cpuPercent, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}
	return cpuPercent, nil
}

func GetCPUUsageTotal() (float64, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	if len(cpuPercent) == 0 {
		return 0, nil
	}
	return cpuPercent[0], nil
}

/*
ret:
 1. percent of usage, 2. total memory, 3. used memory, 4. error
*/
func GetMemUsage() (float64, uint64, uint64, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, 0, err
	}
	return vmem.UsedPercent, vmem.Total, vmem.Used, nil
}

func GetDiskUsage() (float64, uint64, uint64, error) {
	vdisk, err := disk.Usage("/")
	if err != nil {
		return 0, 0, 0, err
	}
	return vdisk.UsedPercent, vdisk.Total, vdisk.Used, nil
}

type CPUInfo struct {
	ModelName string  `json:"model_name"`
	Cores     int32   `json:"cores"`
	Mhz       float64 `json:"mhz"`
}

func GetCPUInfo() (*[]CPUInfo, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	var cpuInfos []CPUInfo
	for _, v := range cpuInfo {
		cpuInfos = append(cpuInfos, CPUInfo{
			ModelName: v.ModelName,
			Cores:     v.Cores,
			Mhz:       v.Mhz,
		})
	}

	return &cpuInfos, nil
}

type ProcessInfo struct {
	Pid     int32   `json:"pid"`
	Name    string  `json:"name"`
	User    string  `json:"user"`
	CPU     float64 `json:"cpu"`
	Memory  uint64  `json:"memory"`
	MemoryP float64 `json:"memory_percent"`
}

func GetProcessInfo() (*[]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var processInfos []ProcessInfo
	for _, v := range processes {
		name, _ := v.Name()
		username, _ := v.Username()
		memory, _ := v.MemoryInfo()
		if memory == nil {
			memory = &process.MemoryInfoStat{}
		}
		processInfos = append(processInfos, ProcessInfo{
			Pid:     v.Pid,
			Name:    name,
			User:    username,
			Memory:  memory.RSS,
			MemoryP: float64(memory.VMS) / float64(memory.RSS),
		})
	}

	return &processInfos, nil
}

/*
ret:
1. recv bytes, 2. sent bytes, 3. error
*/
func GetNetUsage() (uint64, uint64, error) {
	netIO, err := net.IOCounters(true)
	if err != nil {
		return 0, 0, err
	}

	var recv uint64
	var sent uint64
	for _, v := range netIO {
		recv += v.BytesRecv
		sent += v.BytesSent
	}

	return recv, sent, nil
}

func GetNetUsagePercent() (float64, float64, error) {
	recv, sent, err := GetNetUsage()
	if err != nil {
		return 0, 0, err
	}

	total_recv := GetConfigInteger("kisaraClient.network_in")
	total_sent := GetConfigInteger("kisaraClient.network_out")

	recv_percent := float64(recv) / float64(total_recv)
	sent_percent := float64(sent) / float64(total_sent)

	return recv_percent, sent_percent, nil
}

func GetMaxContainer() int {
	return GetConfigInteger("kisaraClient.max_container")
}
