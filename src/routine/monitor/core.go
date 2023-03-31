package routine

import (
	"time"
)

const (
	MONITOR_IP_FREQUENT          = 0x0
	MONITOR_REQUEST              = 0x1
	MONITOR_TRAINING             = 0x2
	MONITOR_RANK                 = 0x3
	MONITOR_MEMCACHE             = 0x4
	MONITOR_DOCKER_FREQUENT      = 0x5
	MONITOR_GAMING               = 0x6
	MONITOR_TRAIN_USER_TAG       = 0x6
	MONITOR_TRADE_CHARGE         = 0x7
	MONITOR_TMPFILE              = 0x8
	MONITOR_VM                   = 0x9
	MONITOR_AWD_STATUS_REFRESHER = 0xA
	MONITOR_USER_LIVING          = 0xB // use to monitor how many users are living
	MONITOR_SCHEDULE_INTERVAL    = 30
)

//监听器
type Monitor struct {
	Type             int
	RefreshInterval  int
	RefreshRemainder int
	Name             string
	Handler          interface{}
	RefreshHandler   func(interface{})
	InitHandler      func(interface{})
}

/*
IP频繁访问频繁检测
首先由IP的大端和小端一起建立索引表，对于每一张表给予“速度”标识，用来记录ip访问频率
对于频率超过阈值的IP，单独建立IP索引表，对单个IP进行记录，单独判断IP访问是否过于频繁
阈值静态设置，管理员可以调整阈值
*/

type FrequentIpListenerChunkDetail struct {
	CauseFault bool //触发风控
	Counter    int  //计数器
}

//ip监控结构
type FrequentIpListenerChunk struct {
	Counter   int
	Threshold int
	Detail    map[int]*FrequentIpListenerChunkDetail
}

type FrequentIpListener struct {
	Map           map[int]*FrequentIpListenerChunk //ip监控池
	MaxTimes      int
	LaunchTimes   int //监控器被调度次数
	ClearInterval int //多少次调度后清除数据
}

//监控器数组
var monitors []*Monitor

//用于停止监控器
var stop_timer_chan chan int

//监听器初始化已完成
var init_finished bool

func getIpIndexByIp(ip int) int {
	return (ip&255)*256 + ((ip & (255 << 24)) >> 24)
}

func (f *FrequentIpListener) CheckAvaliable(ip int) bool {
	if chunk, ok := f.Map[getIpIndexByIp(ip)]; ok {
		if detail, ok := chunk.Detail[ip]; ok {
			return !detail.CauseFault
		}
	}
	return true
}

func (f *FrequentIpListener) Increase(ip int) {
	map_index := getIpIndexByIp(ip)
	if chunk, ok := f.Map[map_index]; ok {
		//看是当前ip是否已经触发警告，有的话就直接处理当前ip
		if detail, ok := chunk.Detail[ip]; ok {
			if detail.CauseFault {
				//检测是否已经触发风险，如果已经触发风险则取消这次increase
				return
			} else if detail.Counter > f.MaxTimes {
				//触发风险，设置CauseFault，在Clear之前这个Ip都废了ovo
				detail.CauseFault = true
			}
			chunk.Detail[ip].Counter++
		} else if chunk.Counter > chunk.Threshold {
			//如果当前ip没有的话就检测ip是否将要触发警告
			var detail FrequentIpListenerChunkDetail
			detail.CauseFault = false
			detail.Counter = chunk.Counter / 2
			chunk.Detail[ip] = &detail
		}
		chunk.Counter++
	} else {
		var chunk FrequentIpListenerChunk
		chunk.Detail = make(map[int]*FrequentIpListenerChunkDetail)
		chunk.Counter = 1
		chunk.Threshold = f.MaxTimes
		f.Map[map_index] = &chunk
	}
}

func (f *FrequentIpListener) Clear() {
	for k := range f.Map {
		delete(f.Map, k)
	}
}

func schedule() {
	for _, monitor := range monitors {
		monitor.RefreshRemainder -= MONITOR_SCHEDULE_INTERVAL
		if monitor.RefreshRemainder <= 0 {
			monitor.RefreshHandler(monitor.Handler)
			monitor.RefreshRemainder += monitor.RefreshInterval
		}
	}
}

func AppendMonitor(monitor *Monitor) {
	monitors = append(monitors, monitor)

	//如果在监听器完成初始化以后再添加心的监听器的话，需要自己执行初始化操作
	if init_finished {
		monitor.RefreshRemainder = monitor.RefreshInterval
		//初始化
		monitor.InitHandler(monitor.Handler)
	}
}

func init() {
	//开启子线程，子线程中开启定时器，定时启动监控器调度程序
	for i := range monitors {
		monitors[i].RefreshRemainder = monitors[i].RefreshInterval
		//初始化
		monitors[i].InitHandler(monitors[i].Handler)
	}
	ticker := time.NewTicker(time.Second * MONITOR_SCHEDULE_INTERVAL)
	go func() {
		for {
			select {
			//接收到定时器信号，开始处理调度程序
			case <-ticker.C:
				schedule()
			case <-stop_timer_chan:
				ticker.Stop()
			}
		}
	}()

	init_finished = true
}

func StopMonitor(code int) {
	stop_timer_chan <- code
}

func TryFrequentIp(ip int) bool {
	handler := monitors[0].Handler.(*FrequentIpListener)
	handler.Increase(ip)
	return handler.CheckAvaliable(ip)
}
