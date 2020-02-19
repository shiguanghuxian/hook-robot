package config

import (
	"os"

	"github.com/naoina/toml"
	"github.com/shiguanghuxian/hook-robot/internal/common"
)

// Config 配置文件
type Config struct {
	Debug  bool     `toml:"debug"`
	Robots []*Robot `toml:"robots"` // 定时列表
}

// Robot 定时任务配置
type Robot struct {
	Type           string   `toml:"type"`            // 定时任务类型 script:脚本 internal: 内部实现
	Target         string   `toml:"target"`          // 目标平台 dingtalk | workweixin
	Name           string   `toml:"name"`            // 机器人名字
	Spec           string   `toml:"spec"`            // 定时
	Cfg            string   `toml:"cfg"`             // 配置内容，格式插件自己定义json script时需要cmd:命令和path脚本路径
	WebHook        string   `toml:"webhook"`         // web hook 地址
	Ats            []string `toml:"ats"`             // 需要@的人列表 手机号
	AtAll          bool     `toml:"at_all"`          // @所有人
	DingtalkSecret string   `toml:"dingtalk_secret"` // 钉钉密钥
}

// NewConfig 初始化一个server配置文件对象
func NewConfig(path string) (cfgChan chan *Config, err error) {
	if path == "" {
		path = common.GetRootDir() + "config/cfg.toml"
	}
	cfgChan = make(chan *Config, 0)
	// 读取配置文件
	cfg, err := readConfFile(path)
	if err != nil {
		return
	}
	go watcher(cfgChan, path)
	go func() {
		cfgChan <- cfg
	}()
	return
}

// ReadConfFile 读取配置文件
func readConfFile(path string) (cfg *Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg = new(Config)
	if err := toml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return
}
