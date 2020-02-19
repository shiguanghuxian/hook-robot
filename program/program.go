package program

import (
	"log"
	"os"
	"time"

	"github.com/robfig/cron"
	"github.com/shiguanghuxian/hook-robot/internal/config"
	"github.com/shiguanghuxian/hook-robot/internal/plugins"
)

var secondParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)

// Program 程序实体
type Program struct {
	cfg    *config.Config
	myCron *cron.Cron
}

// New 创建程序实例
func New() (*Program, error) {
	// 初始化配置文件
	cfgChan, err := config.NewConfig("")
	if err != nil {
		return nil, err
	}

	p := &Program{
		cfg:    <-cfgChan,
		myCron: cron.New(cron.WithParser(secondParser)), // 定时器
	}
	// 配置变化时，软重启服务
	go p.ReloadConfig(cfgChan)

	return p, nil
}

// Run 启动程序
func (p *Program) Run() {
	// js, _ := json.Marshal(p.cfg)
	// log.Println(string(js))

	// 开启每个定时任务
	for _, robot := range p.cfg.Robots {
		robot := robot
		if robot.Spec == "" {
			log.Println("未配置定时执行时间", robot.Type, robot.Name)
			continue
		}
		if robot.WebHook == "" {
			log.Println("未配置web hook url", robot.Type, robot.Name)
			continue
		}
		if robot.Target != plugins.TargetDingtalk && robot.Target != plugins.TargetWorkWeixin {
			log.Println("配置目标平台不支持", robot.Type, robot.Name, robot.Target)
			continue
		}
		onePlugin, ok := plugins.Plugins[robot.Type]
		if !ok || onePlugin == nil {
			log.Println("配置的机器人类型不存在", robot.Type)
			continue
		}
		// 解析定时器格式
		if _, err := secondParser.Parse(robot.Spec); err != nil {
			log.Println("解析定时器格式错误")
			continue
		}
		p.myCron.AddFunc(robot.Spec, func() {
			log.Println("开始执行任务", robot.Type, robot.Name)
			onePlugin.Run(robot)
		})
	}
	// 启动定时任务
	p.myCron.Start()
}

// Stop 程序结束要做的事
func (p *Program) Stop() {
	ctx := p.myCron.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(6 * time.Second):
		log.Println("3秒内停止定时任务未完成")
		os.Exit(1)
	}
}

// ReloadConfig 重新加载配置文件
func (p *Program) ReloadConfig(cfgChan chan *config.Config) {
	for {
		select {
		case <-cfgChan:
			p.Stop()
			p.Run()
		}
	}
}
