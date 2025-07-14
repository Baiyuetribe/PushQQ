package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

var (
	QQClient  *client.QQClient
	logger    = logrus.New()
	dumpspath = "dump"
)

const (
	// 定义颜色代码
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorYellow = "\x1b[33m"
	colorGreen  = "\x1b[32m"
	colorBlue   = "\x1b[34m"
	colorWhite  = "\x1b[37m"
)

const fromProtocol = "Lgr -> "

type protocolLogger struct{}

func (p protocolLogger) Info(format string, arg ...any) {
	logger.Infof(fromProtocol+format, arg...)
}

func (p protocolLogger) Warning(format string, arg ...any) {
	logger.Warnf(fromProtocol+format, arg...)
}

func (p protocolLogger) Debug(format string, arg ...any) {
	logger.Debugf(fromProtocol+format, arg...)
}

func (p protocolLogger) Error(format string, arg ...any) {
	logger.Errorf(fromProtocol+format, arg...)
}

func (p protocolLogger) Dump(data []byte, format string, arg ...any) {
	message := fmt.Sprintf(format, arg...)
	if _, err := os.Stat(dumpspath); err != nil {
		err = os.MkdirAll(dumpspath, 0o755)
		if err != nil {
			logger.Errorf("出现错误 %v. 详细信息转储失败", message)
			return
		}
	}
	dumpFile := path.Join(dumpspath, fmt.Sprintf("%v.dump", time.Now().Unix()))
	logger.Errorf("出现错误 %v. 详细信息已转储至文件 %v 请连同日志提交给开发者处理", message, dumpFile)
	_ = os.WriteFile(dumpFile, data, 0o644)
}

type ColoredFormatter struct{}

func (f *ColoredFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取当前时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 根据日志级别设置相应的颜色
	var levelColor string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = colorBlue
	case logrus.InfoLevel:
		levelColor = colorGreen
	case logrus.WarnLevel:
		levelColor = colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = colorRed
	default:
		levelColor = colorWhite
	}

	return utils.S2B(fmt.Sprintf("[%s] [%s%s%s]: %s\n",
		timestamp, levelColor, strings.ToUpper(entry.Level.String()), colorReset, entry.Message)), nil
}

func init() {
	logger.SetLevel(logrus.PanicLevel) // PanicLevel 时，只会输出 Panic 级别的日志；最终环境使用
	logger.SetFormatter(&ColoredFormatter{})
	logger.SetOutput(colorable.NewColorableStdout())
}

// InitQQBot 初始化QQ机器人
func InitQQBot() error {
	// 使用特定的协议版本
	appInfo := auth.AppList["linux"]["3.2.15-30366"]
	// 创建设备信息
	deviceInfo := &auth.DeviceInfo{
		GUID:          "cfcd208495d565ef66e7dff9f967642a",
		DeviceName:    "Lagrange-DCFCD07E",
		SystemKernel:  "Windows 10.0.22631",
		KernelVersion: "10.0.22631",
	}

	// 创建qqclient实例
	QQClient = client.NewClient(0, "")
	// 设置qqclient的logger
	QQClient.SetLogger(protocolLogger{})
	// 使用协议版本
	QQClient.UseVersion(appInfo)
	// 添加signserver，注意要和appinfo版本匹配
	QQClient.AddSignServer("https://sign.lagrangecore.org/api/sign/30366")
	// 使用设备信息
	QQClient.UseDevice(deviceInfo)

	// 从保存的sig.bin文件读取登录信息
	data, err := os.ReadFile("sig.bin")
	if err != nil {
		logrus.Warnln("read sig error:", err)
	} else {
		// 将登录信息反序列化
		sig, err := auth.UnmarshalSigInfo(data, true)
		if err != nil {
			logrus.Warnln("load sig error:", err)
		} else {
			// 如果登录信息有效，则使用登录信息登录
			QQClient.UseSig(sig)
		}
	}

	// 订阅群消息事件 --- 打印全部消息
	QQClient.GroupMessageEvent.Subscribe(func(client *client.QQClient, event *message.GroupMessage) {
		// fmt.Println(event.ToString()) // 打印群消息内容
		// fmt.Println(event.Sender.Uin) // 打印发送者的QQ号
		// fmt.Println(event.GroupUin)   // 打印群号
		// msg := message.NewText("Hello, this is a test message!")
		// _, err := client.SendGroupMessage(event.GroupUin, []message.IMessageElement{msg})
		// if err != nil {
		// 	return
		// }
	})
	// 订阅私聊消息事件 --- 独立QQ对话
	QQClient.PrivateMessageEvent.Subscribe(func(client *client.QQClient, event *message.PrivateMessage) {
		// 这段代码会将群聊收到的消息打印出来
		// fmt.Println(event.ToString())
		// fmt.Println(event.Sender.UID)
		// fmt.Println(event.Sender.Uin)

		if event.ToString() == "ping" {
			msg := message.NewText("pong")
			_, err := client.SendPrivateMessage(event.Sender.Uin, []message.IMessageElement{msg})
			if err != nil {
				return
			}
		}
	})

	QQClient.DisconnectedEvent.Subscribe(func(client *client.QQClient, event *client.DisconnectedEvent) {
		logger.Infof("连接已断开：%v", event.Message)
	})

	// 登录处理
	err = loginBot(QQClient, false)
	if err != nil {
		logger.Errorln("login err:", err)
		return err
	}
	logger.Infoln("login successed")

	// 设置优雅关闭
	go func() {
		mc := make(chan os.Signal, 2)
		signal.Notify(mc, os.Interrupt, syscall.SIGTERM)
		<-mc

		// 序列化登录信息以便下次使用
		data, err := QQClient.Sig().Marshal()
		if err != nil {
			logger.Errorln("marshal sig.bin err:", err)
			return
		}
		err = os.WriteFile("sig.bin", data, 0644)
		if err != nil {
			logger.Errorln("write sig.bin err:", err)
			return
		}
		logger.Infoln("sig saved into sig.bin")

		QQClient.Release()
		os.Exit(0)
	}()

	return nil
}

func loginBot(c *client.QQClient, passwordLogin bool) error {
	logger.Info("login with password")
	// 如果登录信息存在，可以使用fastlogin
	err := c.FastLogin()
	if err == nil {
		return nil
	}

	if passwordLogin {
		// 密码登录，目前无法使用
		ret, err := c.PasswordLogin()
		for {
			if err != nil {
				logger.Errorf("密码登录失败: %s", err)
				break
			}
			if ret.Success {
				return nil
			}
			switch ret.Error {
			case client.SliderNeededError:
				logger.Warnln("captcha verification required")
				logger.Warnln(ret.VerifyURL)
				aid := strings.Split(strings.Split(ret.VerifyURL, "sid=")[1], "&")[0]
				logger.Warnln("ticket?->")
				ticket := utils.ReadLine()
				logger.Warnln("rand_str?->")
				randStr := utils.ReadLine()
				ret, err = c.SubmitCaptcha(ticket, randStr, aid)
				continue
			case client.UnsafeDeviceError:
				vf, err := c.GetNewDeviceVerifyURL()
				if err != nil {
					return err
				}
				logger.Infoln(vf)
				err = c.NewDeviceVerify(vf)
				if err != nil {
					return err
				}
			default:
				logger.Errorf("Unhandled exception raised: %s", ret.ErrorMessage)
			}
		}
	}
	logger.Infoln("login with qrcode")

	// 扫码登录流程
	// 首先获取二维码
	png, _, err := c.FetchQRCodeDefault()
	if err != nil {
		return err
	}
	qrcodePath := "qrcode.png"
	// 保存到本地以供扫码
	err = os.WriteFile(qrcodePath, png, 0666)
	if err != nil {
		return err
	}
	fmt.Printf("二维码已保存，请扫码以登录: %s\n", qrcodePath)
	for {
		// 轮询二维码扫描结果
		retCode, err := c.GetQRCodeResult()
		if err != nil {
			log.Println(err)
			return err
		}
		// 等待扫码
		if retCode.Waitable() {
			time.Sleep(3 * time.Second)
			continue
		}
		if !retCode.Success() {
			return errors.New(retCode.Name())
		}
		break
	}
	// 扫码完成后就可以进行登录
	_, err = c.QRCodeLogin()
	return err
}
