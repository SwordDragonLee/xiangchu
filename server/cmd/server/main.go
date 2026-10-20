// 乡厨系统 API（Swagger 文档锚点，task server:swagger 重新生成）
//
//	@Title			乡厨系统 API
//	@Version		1.0
//	@Description	乡厨宴席服务平台后端。统一返回 HTTP 200 + 业务码信封 {code, msg, data}：code=0 成功，40xxx 业务错误（40000 参数 / 40100 未登录 / 40300 无权限 / 40400 不存在 / 40900 状态冲突），50000 服务错误。金额单位为分，费率单位为基点 BP（3%=300）。user 组接口同时接受 user/chef/supplier 三种 token。
//	@Schemes		http
//	@BasePath		/
//	@tag.name		用户域
//	@tag.description	登录/验证码、资料、地址簿、优惠券
//	@tag.name		乡厨域
//	@tag.description	乡厨列表/详情/档期/套餐/入驻
//	@tag.name		订单域
//	@tag.description	宴席订单状态机（0待支付→1待接单→2已接单→3备菜中→4服务中→5已完成，6已取消/7已拒绝）与评价售后
//	@tag.name		采购供应域
//	@tag.description	食材市场、采购单、供应商履约
//	@tag.name		财务域
//	@tag.description	支付回调/模拟支付、结算、提现
//	@tag.name		管理端
//	@tag.description	后台登录、审核、监控、统计、财务配置
//	@SecurityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				粘贴 "Bearer {token}"（必须含 Bearer 前缀和空格，后端校验该前缀）。token 由各登录接口返回。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xiangchu/server/internal/bootstrap"
	"github.com/xiangchu/server/internal/config"
)

func main() {
	cfgPath := flag.String("c", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "加载配置失败:", err)
		os.Exit(1)
	}
	if err := bootstrap.Run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "服务启动失败:", err)
		os.Exit(1)
	}
}
