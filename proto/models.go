package proto

// Config 用户提交的任务配置
type Config struct {
	ID               string      // 唯一标识，后台自动生成
	Name             string      // 配置名
	TargetService    string      // 被调服务
	Task                         // 任务
	Trigger                      // 触发条件
	Dependencies     Dependency  // 任务依赖
	RetryStrategy                // 失败处理策略
	ShardingStrategy             // 分片策略
	ShardingResults  []*Sharding // 静态分片结果
}

type Task struct {
	Type   string            // 任务类型
	URI    string            // URI
	Body   string            // 任务体
	Header map[string]string // 任务头
}

type Trigger struct {
}

type Dependency struct {
	Nodes map[string]string //涉及的所有任务id
	Links []Edge            //任务依赖关系
}

type NodeEntity struct {
	Id     string
	status string
}

type Edge struct {
	From string //前置条件
	To   string //后置任务
}

type RetryStrategy struct {
}

type ShardingStrategy struct {
	ShardingType  string // 分片类型：静态/动态
	ShardingCount int    // 分片数量
	DefaultCount  int    // 默认执行器数量
	ParameterRole string // 分片规则 0=A，1=B
}

type Sharding struct {
	ShardingItem int    // 分片序号
	Parameter    string // 分片参数
	Instance     string // 执行实例
}
