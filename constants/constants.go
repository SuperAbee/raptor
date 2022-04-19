package constants

const (
	NACOS_GROUP               = "raptor"
	K8S_NAMESPACE             = "default"
	K8S_CONFIGMAP_CONTENT_KEY = "content"
	K8S_GROUP_LABEL           = "raptor"
	K8S_APP_LABEL             = "raptor"
	TASK_UNEXECUTED           = "unexecuted"
	TASK_RUNNING              = "running"
	TASK_COMPLETED            = "completed"
	TASK_FAIL                 = "fail"
	SINGLE_JOB                = "job"
	DEPENDENCE_JOB            = "deJob"
	DEPENDENCE_SUB_JOB        = "deSubJob"
	JOB_GROUP                 = "Job"
	DEPENDENCE_GROUP          = "Dependence"
	TASK_SHARDING_GROUP       = "ShardingInfo"
	JOB_INSTANCE_GROUP        = "JobInstance"
	DEPENDENCE_INSTANCE_GROUP = "DependenceInstance"
	DEPENDENCCE_DAG           = "DependenceDAG"
)
