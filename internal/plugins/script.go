package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/shiguanghuxian/hook-robot/internal/config"
)

// ScriptPlugin 执行脚本获取消息发送数据
type ScriptPlugin struct {
}

func init() {
	registerPlugin("script", new(ScriptPlugin))
}

// Run 运行脚本插件
func (p *ScriptPlugin) Run(cfg *config.Robot) error {
	if cfg == nil {
		return errors.New("参数为nil")
	}
	// 解析cfg附件参数
	scriptCfg := make(map[string]string, 0)
	err := json.Unmarshal([]byte(cfg.Cfg), &scriptCfg)
	if err != nil {
		log.Println("解析脚本配置错误", err)
		return err
	}
	if scriptCfg["cmd"] == "" || scriptCfg["path"] == "" {
		log.Println("配置中缺少cmd或path", cfg.Cfg)
		return errors.New("配置中缺少cmd或path")
	}
	args := make([]string, 0)
	args = append(args, scriptCfg["path"])
	// 初始化Cmd
	command := exec.Command(scriptCfg["cmd"], args...)
	// 输出数据
	var out bytes.Buffer
	command.Stdout = &out
	// 运行脚本
	err = command.Start()
	if err != nil {
		log.Println("执行脚本错误1", err)
		return err
	}
	// log.Println("脚本进程PID", command.Process.Pid)
	err = command.Wait() //等待执行完成
	if err != nil {
		log.Println("执行脚本错误2", err, string(out.Bytes()))
		return err
	}
	// 根据平台发送数据
	fmt.Println("开始发送web hook")
	err = SendWebHookByCfg(cfg, out.Bytes())
	if err != nil {
		log.Println("发送web hook失败")
		return err
	}
	fmt.Println("发送web hook成功")
	return nil
}
