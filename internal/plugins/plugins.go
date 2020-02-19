package plugins

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/shiguanghuxian/hook-robot/internal/config"
)

// Plugin 插件接口
type Plugin interface {
	Run(cfg *config.Robot) error
}

// TargetType 接收目标
type TargetType string

const (
	TargetDingtalk   = "dingtalk"   // 钉钉
	TargetWorkWeixin = "workweixin" // 企业微信
)

var (
	// Plugins 所有插件
	Plugins map[string]Plugin
)

func init() {
	Plugins = make(map[string]Plugin, 0)
}

// 注册插件
func registerPlugin(name string, p Plugin) {
	Plugins[name] = p
}

// SendWebHook 发送数据到web hook
func SendWebHook(hookUrl string, payload []byte) error {
	if hookUrl == "" {
		return errors.New("web hook url 不能为空")
	}
	req, err := http.NewRequest("POST", hookUrl, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// 打印日志
	body, _ := ioutil.ReadAll(res.Body)
	log.Println("发送web hook结果", hookUrl, string(body))
	return nil
}

// SendWebHookByCfg 发送web hook参数为配置文件
func SendWebHookByCfg(cfg *config.Robot, data interface{}) error {
	payload := make(map[string]interface{}, 0)
	switch data.(type) {
	case string:
		body := []byte(data.(string))
		json.Unmarshal(body, &payload)
	case []byte:
		body := data.([]byte)
		json.Unmarshal(body, &payload)
	case map[string]interface{}:
		payload = data.(map[string]interface{})
	default:
		return errors.New("不支持的data类型")
	}

	hookUrl := cfg.WebHook
	dingtalkSecret := cfg.DingtalkSecret
	if cfg.Target == TargetDingtalk {
		// 计算钉钉签名
		if dingtalkSecret != "" && cfg.Target == TargetDingtalk {
			timestamp := time.Now().UnixNano() / 1e6
			stringToSign := fmt.Sprintf("%d\n%s", timestamp, dingtalkSecret)
			h := hmac.New(sha256.New, []byte(dingtalkSecret))
			io.WriteString(h, stringToSign)
			hmacCode := h.Sum(nil)
			sign := url.QueryEscape(base64.StdEncoding.EncodeToString(hmacCode))
			// fmt.Println(sign)
			// 拼接url参数
			hookUrl = fmt.Sprintf("%s&timestamp=%d&sign=%s", hookUrl, timestamp, sign)
		}
		// 追加at数据
		msgtype, ok := payload["msgtype"].(string)
		if ok == false {
			return errors.New("消息类型不能为空")
		}
		switch msgtype {
		case "text", "markdown":
			if at, ok := payload["at"].(map[string]interface{}); ok {
				atMobiles, ok := at["atMobiles"].([]string)
				if ok {
					atMobiles = make([]string, 0)
				}
				if len(cfg.Ats) > 0 {
					atMobiles = append(atMobiles, cfg.Ats...)
				}
				at["atMobiles"] = atMobiles
				at["isAtAll"] = cfg.AtAll
				payload["at"] = at
			} else {
				at = make(map[string]interface{}, 0)
				at["atMobiles"] = cfg.Ats
				at["isAtAll"] = cfg.AtAll
				payload["at"] = at
			}
		}
	} else if cfg.Target == TargetWorkWeixin {
		if len(cfg.Ats) > 0 || cfg.AtAll == true {
			// 追加at 支持部分消息类型
			if payload["msgtype"] == "text" {
				if text, ok := payload["text"].(map[string]interface{}); ok {
					mobileList, ok := text["mentioned_mobile_list"].([]string)
					if !ok {
						mobileList = make([]string, 0)
					}
					if len(cfg.Ats) > 0 {
						mobileList = append(mobileList, cfg.Ats...)
					}
					if cfg.AtAll == true {
						mobileList = append(mobileList, "@all")
					}
					text["mentioned_mobile_list"] = mobileList
					payload["text"] = text
				}
			}
		}
	}
	// 请求web hook
	js, _ := json.Marshal(payload)
	log.Println(string(js))
	return SendWebHook(hookUrl, js)
}
