package dependence

import (
	"errors"
	"math"
	"math/rand"
	"raptor/proto"
)

const MAX_LEVEL = 16 // 最大索引层级限制

// 不支持重复元素
// 支持相同的score
type SkipList struct {
	Head   *skipListNode
	LevelN int // 包含原始链表的层数
	Length int // 长度
}
type skipListNode struct {
	Val   *proto.JobInstance
	Level int // 该节点的最高索引层是多少
	Score int64
	Next  []*skipListNode // key是层高
}

func NewSkipList() *SkipList {
	return &SkipList{
		Head:   newSkipListNode(nil, math.MinInt64, MAX_LEVEL),
		LevelN: 1,
		Length: 0,
	}
}

func newSkipListNode(val *proto.JobInstance, score int64, level int) *skipListNode {
	return &skipListNode{
		Val:   val,
		Level: level,
		Score: score,
		Next:  make([]*skipListNode, level, level),
	}
}

//获取即将执行的任务
func (sl *SkipList) getLessThan(score int64) map[*proto.JobInstance]int64 {
	maps := make(map[*proto.JobInstance]int64)
	if sl.LevelN == 0 {
		return maps
	}
	cur := sl.Head
	for cur.Next[0] != nil {
		if cur.Next[0].Score <= score {
			maps[cur.Next[0].Val] = cur.Next[0].Score
		} else {
			break
		}
		cur = cur.Next[0]
	}
	//刷新跳表，防止重复执行
	for k, v := range maps {
		sl.Delete(k, v)
	}
	return maps
}

func (sl *SkipList) Insert(val *proto.JobInstance, score int64) (bool, error) {
	if val == nil {
		return false, errors.New("can't insert nil value to skiplist")
	}

	cur := sl.Head
	update := [MAX_LEVEL]*skipListNode{} // 记录在每一层的插入位置，value保存哨兵结点
	k := MAX_LEVEL - 1
	// 从最高层的索引开始查找插入位置，逐级向下比较，最后插入到原始链表也就是第0级
	for ; k >= 0; k-- {
		for cur.Next[k] != nil {
			if cur.Next[k].Val == val {
				return false, errors.New("can't insert repeatable value to skiplist")
			}
			if cur.Next[k].Score > score {
				update[k] = cur
				break
			}
			cur = cur.Next[k]
		}
		// 如果待插入元素的优先级最大，哨兵节点就是最后一个元素
		if cur.Next[k] == nil {
			update[k] = cur
		}
	}

	randomLevel := sl.getRandomLevel()
	newNode := newSkipListNode(val, score, randomLevel)

	// 插入元素
	for i := randomLevel - 1; i >= 0; i-- {
		newNode.Next[i] = update[i].Next[i]
		update[i].Next[i] = newNode
	}
	if randomLevel > sl.LevelN {
		sl.LevelN = randomLevel
	}
	sl.Length++

	return true, nil
}

// skiplist在插入元素时需要维护索引，生成一个随机值，将元素插入到第1-k级索引中
func (sl *SkipList) getRandomLevel() int {
	level := 1
	for i := 1; i < MAX_LEVEL; i++ {
		if rand.Int31()%7 == 1 {
			level++
		}
	}
	return level
}

func (sl *SkipList) Find(v *proto.JobInstance, score int64) *skipListNode {
	if v == nil || sl.Length == 0 {
		return nil
	}
	cur := sl.Head
	for i := sl.LevelN - 1; i >= 0; i-- {
		if cur.Next[i] != nil {
			if cur.Next[i].Val == v && cur.Next[i].Score == score {
				return cur.Next[i]
			} else if cur.Next[i].Score >= score {
				continue
			}
			cur = cur.Next[i]
		}
	}
	// 如果没有找到该元素，这时cur是原始链表中，score相同的第一个元素，向后查找
	for cur.Next[0].Score <= score {
		if cur.Next[0].Val == v && cur.Next[0].Score == score {
			return cur.Next[0]
		}
		cur = cur.Next[0]
	}

	return nil
}
func (sl *SkipList) Delete(v *proto.JobInstance, score int64) bool {
	if v == nil {
		return false
	}
	cur := sl.Head
	// 记录每一层待删除数据的前驱结点
	// 如果某些层没有待删除数据，那么update[i]为空
	// 如果待删除数据不存在，那么update[i]也为空
	update := [MAX_LEVEL]*skipListNode{}
	for i := sl.LevelN - 1; i >= 0; i-- {
		for cur.Next[i] != nil && cur.Next[i].Score <= score {
			if cur.Next[i].Score == score && cur.Next[i].Val == v {
				update[i] = cur
				break
			}
			cur = cur.Next[i]
		}
	}
	// 删除节点
	for i := sl.LevelN - 1; i >= 0; i-- {
		if update[i] == nil {
			continue
		}
		// 如果该层中，删除节点是第一个节点且没有下一个节点，直接降低索引层（只有最高层会出现这种情况）
		if update[i] == sl.Head && update[i].Next[i].Next[i] == nil {
			sl.LevelN = i
			continue
		}
		update[i].Next[i] = update[i].Next[i].Next[i]
	}

	sl.Length--

	return true
}
