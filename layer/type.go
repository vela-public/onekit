package layer

import (
	"net"
	"time"
)

// Ident 节点身份信息 重载了tunnel.Ident 方便接口统一
type Ident struct {
	Inet       net.IP        `json:"inet"`       // 内网出口 IP
	MAC        string        `json:"mac"`        // 出口 IP 所在网卡的 MAC 地址
	CPU        int           `json:"cpu"`        // CPU 核心数
	PID        int           `json:"pid"`        // 进程 PID
	Workdir    string        `json:"workdir"`    // 工作目录
	Executable string        `json:"executable"` // 执行路径
	Username   string        `json:"username"`   // 当前操作系统用户名
	Hostname   string        `json:"hostname"`   // 主机名
	Interval   time.Duration `json:"interval"`   // 心跳间隔，如果中心端 3 倍心跳仍未收到任何消息，中心端强制断开该连接
	TimeAt     time.Time     `json:"time_at"`    // agent 当前时间
	Goos       string        `json:"goos"`       // 操作系统
	Arch       string        `json:"arch"`       // 操作系统架构
	Semver     string        `json:"semver"`     // 节点版本
	Unload     bool          `json:"unload"`     // 是否开启静默模式，仅在新注册节点时有效
	Unstable   bool          `json:"unstable"`   // 不稳定版本
	Customized string        `json:"customized"` // 定制版本
}

// MHide 节点身份信息 重载了definition.MHide 方便接口统一
type MHide struct {
	// Servername 服务端域名。
	// 此处有两个作用：
	// 		TLS 连接下，用于证书认证校验。
	//
	Servername string `json:"servername"`

	// Addrs 代理节点地址，告诉 agent 客户端连接哪台代理节点。
	Addrs []string `json:"addrs"`

	// Semver agent 二进制发行版本。
	Semver string `json:"semver"`

	// Hash 文件原始哈希，不计算隐写数据。
	Hash string `json:"hash"`

	// Size 文件原始大小，不计算隐写数据。
	Size int64 `json:"size"`

	// Tags 首次下载时的隐写标签，
	// 例如：首次部署时，
	Tags []string `json:"tags"`

	// Goos 操作系统。
	Goos string `json:"goos"`

	// Arch CPU 架构。
	Arch string `json:"arch"`

	// Unload 是否开启静默模式，仅对新注册上线的节点有效。
	Unload bool `json:"unload"`

	// Unstable 是否不稳定版本。
	Unstable bool `json:"unstable"`

	// Customized 定制版本标记，为空代表标准版或叫通用版。
	Customized string `json:"customized"`

	// DownloadAt 下载时间。
	DownloadAt time.Time `json:"download_at"`

	// VIP 代理节点公网地址。
	//
	// Deprecated: use Addrs.
	VIP []string `json:"vip"`

	// LAN 代理节点内网地址。
	//
	// Deprecated: use Addrs.
	LAN []string `json:"lan"`

	// Edition 版本号。
	//
	// Deprecated: use Semver.
	Edition string `json:"edition"`
}
