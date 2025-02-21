package treekit

type TaskConfig struct {
	// task ID
	ID int64 `json:"id"`

	//每条任务可以被多次触发执行，每次执行时，都会生成一个唯一执行ID
	ExecID int64 `json:"exec_id"`

	//任务名称
	Name string `json:"name"`

	//这里是任务说明简介，帮助用户理解该程序。
	Intro string `json:"intro"`

	// source
	Code string `json:"code"`

	// task code hash
	CodeSha1 string `json:"code_sha1"`

	//timeout
	Timeout int64 `json:"timeout"`
}
