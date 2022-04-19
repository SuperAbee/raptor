package proto

// Config 用户提交的任务配置
type Config struct {
	ID               string           // 唯一标识，后台自动生成
	Name             string           // 配置名
	Description      string           // 任务描述
	TargetService    string           // 被调服务
	Cron             string           // 执行时间
	Task             Task             // 任务请求
	Trigger          Trigger          // 触发条件
	Dependencies     Dependency       // 任务依赖
	RetryStrategy    RetryStrategy    // 失败处理策略
	RetryCount       int              // 重试次数
	ShardingStrategy ShardingStrategy // 分片策略
	PreFilter        []string         // 前置过滤器
	PostFilter       []string         // 后置过滤器
	Executor         string           // 执行器
}

// JobInstance 用于执行的任务实例
type JobInstance struct {
	Config       Config
	ExecuteTime  int64
	StartTime    int64
	EndTime      int64
	ID           string
	PreId        string
	Type         string
	IsMaster     bool
	ExecuteCount int
	Status       string
	Extra        map[string]interface{}
}
type Task struct {
	Type   string            // 请求类型
	URI    string            // URI
	Body   string            // 请求体
	Header map[string]string // 请求头
}

type Trigger struct {
}

type Dependency struct {
	Nodes map[string]NodeInfo //涉及的所有任务id
	Links []Edge              //任务依赖关系
}

type NodeInfo struct {
	NodeName   string
	JobId      string
	InstanceID string
	Status     string
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
	Ip           string // 执行实例ip
	Port         uint64 // 执行实例port
}
