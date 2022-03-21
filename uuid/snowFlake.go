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
	now := time.Since(sf.epoch).Nanoseconds() / 1000000
	if now == sf.lastTime {
		sf.sequence = (sf.sequence + 1) & SeqNumMax
		if sf.sequence == 0 {
			for now <= sf.lastTime {
				now = time.Since(sf.epoch).Nanoseconds() / 1000000
			}
		}
	} else {
		sf.sequence = 0
	}
	sf.lastTime = now
	id := (now << timeBits) |
		(sf.machineID << SeqNumBits) |
		sf.sequence
	sf.mu.Unlock()
	return id
}
