package handler

import (
	"push_qq/core"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/gofiber/fiber/v2"
)

func PostMsg(c *fiber.Ctx) error {
	// jwt信息校验
	input := struct {
		Method string `json:"method"` // 方法名 qq或群
		Uid    string `json:"uid"`    // 用户 对应QQ号或群号
		Msg    string `json:"msg"`    // 消息内容
	}{}
	if err := c.BodyParser(&input); err != nil || len(input.Msg) == 0 {
		return c.JSON(fiber.Map{"status": "error", "msg": "数据请求错误"})
	}

	// 检查QQ客户端是否已初始化
	if core.QQClient == nil {
		return c.JSON(fiber.Map{"status": "error", "msg": "QQ客户端未初始化"})
	}

	// 发送消息
	switch input.Method {
	case "qq":
		// 发送QQ消息
		uin, err := strconv.ParseUint(input.Uid, 10, 64)
		if err != nil {
			return c.JSON(fiber.Map{"status": "error", "msg": "QQ号格式错误"})
		}

		msg := message.NewText(input.Msg)
		_, err = core.QQClient.SendPrivateMessage(uint32(uin), []message.IMessageElement{msg})
		if err != nil {
			return c.JSON(fiber.Map{"status": "error", "msg": "发送私聊消息失败: " + err.Error()})
		}

	case "group":
		// 发送群消息
		groupUin, err := strconv.ParseUint(input.Uid, 10, 64)
		if err != nil {
			return c.JSON(fiber.Map{"status": "error", "msg": "群号格式错误"})
		}

		msg := message.NewText(input.Msg)
		_, err = core.QQClient.SendGroupMessage(uint32(groupUin), []message.IMessageElement{msg})
		if err != nil {
			return c.JSON(fiber.Map{"status": "error", "msg": "发送群消息失败: " + err.Error()})
		}

	default:
		return c.JSON(fiber.Map{"status": "error", "msg": "未知方法"})
	}

	return c.JSON(fiber.Map{"status": "success", "msg": "消息发送成功"})
}
