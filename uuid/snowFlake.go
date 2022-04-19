package uuid

import (
	"fmt"
	"sync"
	"time"
)

var (
	//序列号范围
	SeqNumBits uint8 = 12
	SeqNumMax  int64 = -1 ^ (-1 << SeqNumBits)
	//机器范围
	NodeBits uint8 = 10
	NodeMax  int64 = -1 ^ (-1 << NodeBits)
	NodeMask int64 = NodeMax << SeqNumBits
	Epoch    int64 = 1288834974657
	//时间偏移量
	timeBits uint8 = SeqNumBits + NodeBits
)

type SnowFlakeUUID struct {
	mu        sync.Mutex
	machineID int64
	epoch     time.Time
	lastTime  int64
	sequence  int64
}

func NewSnowFlakeUUID(machineID int64) (*SnowFlakeUUID, error) {
	if machineID > NodeMax || machineID <= 0 {
		return nil, fmt.Errorf("machineID too large or less zero")
	}
	n := SnowFlakeUUID{}
	n.machineID = machineID
	var curTime = time.Now()
	n.epoch = curTime.Add(time.Unix(Epoch/1000, (Epoch%1000)*1000000).Sub(curTime))
	return &n, nil
}

func (sf *SnowFlakeUUID) GenerateID() int64 {
	sf.mu.Lock()
	// 获取当前时间戳
	now := time.Since(sf.epoch).Nanoseconds() / 1000000
	// 与上次获取Id的时间进行对比
	if now == sf.lastTime {
		// 在同一毫秒中获取Id，计算当前序列号
		sf.sequence = (sf.sequence + 1) & SeqNumMax

		// 如果序列号达到最大值，则等到下一毫秒进行获取
		if sf.sequence == 0 {
			for now <= sf.lastTime {
				now = time.Since(sf.epoch).Nanoseconds() / 1000000
			}
		}
	} else {
		sf.sequence = 0
	}
	sf.lastTime = now
	// 将Id通过位运算进行拼接
	id := (now << timeBits) |
		(sf.machineID << SeqNumBits) |
		sf.sequence
	sf.mu.Unlock()
	return id
}
