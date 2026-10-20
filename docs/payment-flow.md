# 支付全流程（微信 JSAPI）

本文描述宴席订单、食材采购单从「用户点支付」到「后端确认入账」的完整链路，覆盖前端、后端、微信三方。对应代码为 v1 现状。

## 一、参与方

| 角色 | 说明 |
|---|---|
| 前端 | user-mp / chef-mp 小程序，负责发起支付与展示结果 |
| 后端 | server（Gin），支付单账本持有者，唯一可信入账方 |
| 微信支付 | 收银台（用户输密码）+ 记账系统（扣款后异步回调） |
| 基础设施 | MySQL `payments` 表、RabbitMQ `pay.success` 事件 |

## 二、时序总览

```
小程序                后端(server)                    微信支付
  │ ① 请求支付参数        │                              │
  │   POST /orders/:id/pay-balance（采购单为创建时返回）      │
  │──────────────────> │                              │
  │                    │ ② payments 插入 status=0      │
  │                    │ ③ 统一下单 /v3/pay/transactions/jsapi
  │                    │───────────────────────────>  │
  │                    │          prepay_id            │
  │                    │<───────────────────────────   │
  │ ④ 返回调起六件套       │   （商户私钥签名 paySign）         │
  │<─────────────────── │                              │
  │ ⑤ uni.requestPayment（直连微信收银台，不经后端）           │
  │─────────────────────────────────────────────────>  │
  │                    │       ⑥ 用户付款，微信记账        │
  │                    │ <──── POST notify_url ──────   │
  │                    │ ⑦ 验签 → 解密 → 乐观锁入账        │
  │                    │ ───── {"code":"SUCCESS"} ───>  │
  │                    │ ⑧ 发布 MQ pay.success 事件      │
  │ ⑨ 跳详情页重新拉取 → 显示已支付                           │
```

## 三、分步详解

### 1. 前端：请求支付参数

入口统一收口在 `doPay` 模式（以订单详情页为例）：

```ts
// web/apps/user-mp/src/pages/order/detail.vue
function doPay(pay) {
  pay()                                  // POST /orders/:id/pay-balance
    .then((res) => handlePayResult(res)) // 拉起收银台
    .then(() => load())                  // 重新拉订单
}
```

- 宴席订单：定金/尾款/全款统一走 `POST /orders/:id/pay-balance`，返回 `PayResult`
- 食材采购：`POST /purchases` 创建成功时直接返回 `payment_no + params`

前端对后端说的唯一一句话是「给我调起支付的参数」。**前端没有任何「我已支付」的上报通道**——它没有资格宣布支付结果。

### 2. 后端：创建支付单

`internal/modules/pay/service.go` Create：

- `orderid.GenNo("PAY")` 生成商户单号 `out_trade_no`（全项目唯一）
- `payments` 表插入一条 `status=0`（待支付）记录，含 `biz_type/biz_id/amount/expire_time`
- 这条记录是整个支付的钱账账本；微信侧只认 `out_trade_no + mchid` 组合

### 3. 后端：统一下单 + 签名

`internal/infra/wxpay/jsapi.go` JSAPI：

- `POST https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi`
- 请求体：`appid / mchid / out_trade_no / notify_url / amount.total(分) / payer.openid`
- 微信返回 `prepay_id`（预支付单，2 小时有效）
- 用商户 API 私钥对 `appid\nts\nonce\nprepay_id=xxx\n` 五行做 SHA256-RSA 签名 → `paySign`
- 返回前端六件套：`app_id / time_stamp / nonce_str / package(prepay_id=...) / pay_sign`（附 `payment_no`）

`notify_url` 按单携带：下单时带上当前环境的回调地址，测试与生产互不干扰。

### 4. 前端：拉起收银台

```ts
// web/apps/user-mp/src/utils/pay.ts
uni.requestPayment({ provider: 'wxpay', timeStamp, nonceStr, package, signType: 'RSA', paySign })
```

- 此步通信直连微信，不经过后端
- `success` 只代表「用户完成了收银台操作」，**不是入账证明**（可能掉单）

### 5. 微信：扣款与异步回调

- 用户输密码 → 微信记账系统扣款，资金结算到商户账户
- 微信服务器向下单时登记的 `notify_url` 发起 HTTPS POST：
  - 报文整体 AES-GCM 加密，并带微信平台私钥签名
  - 未收到 `SUCCESS` 应答会按 15s/15s/30s/3m/10m… 衰减重试，最长 24 小时

### 6. 后端：验签入账（唯一真相）

`internal/modules/pay/handler.go:62` → `service.go` HandleWxNotify：

1. 路由 `POST /pay/notify/wx`——**不走 JWT 鉴权**，微信的身份证明是平台签名
2. `ParseNotify`：微信平台证书验签 + 解密报文
3. `TradeState != "SUCCESS"` 拒收
4. 按 `out_trade_no` 反查支付单
5. `MarkPaid` 乐观锁入账（`pay/repo.go:44`）：

```sql
UPDATE payments SET status=1, wx_transaction_id=?, pay_time=?
WHERE id=? AND status=0
```

   - `RowsAffected=1`：本次调用者赢得入账权
   - `RowsAffected=0`：已被入账（回调重放/并发），安全跳过——**幂等的关键**
6. 应答 `{"code":"SUCCESS"}`，微信停止重发

### 7. 后端：广播领域事件

`MarkPaid` 成功的瞬间发布 `pay.success` 事件（RabbitMQ EventExchange）。消费者（`internal/bootstrap/bootstrap.go:76-79`）：

| 消费方 | 动作 |
|---|---|
| order 域 | 订单状态推进（待支付 → 已预订） |
| food 域 | 采购单状态推进 |
| 结算域 | 生成分账结算单（净额 = 金额 − 佣金，费率存 `system_settings` 可调） |
| notify | 日志记录 |

支付域与业务域解耦：微信回调只进支付域，业务域各自订阅事件认领后续。

### 8. 前端：拉取最终状态

- `requestPayment` success 回来即跳订单/采购单详情页
- 详情页 `onShow` 重新 GET，拉到的即是回调入账后的真相
- `success` 与回调到达之间可能差零点几秒，页面短暂显示待支付属正常

## 四、支付单状态机

常量定义 `internal/modules/pay/model.go:13-16`：

| status | 含义 | 流转 |
|---|---|---|
| 0 | 待支付 | 创建时写入；唯一可入账状态 |
| 1 | 已支付 | 乐观锁 0→1（仅回调触发） |
| 2 | 退款中 | CAS 1→2，发起退款 |
| 3 | 已退款 | 退款完成 |

超时未支付的支付单停留在 0（`expire_time` 记录过期时间，微信侧 prepay_id 2 小时失效）；主动关单与查单兜底为待接入项。

## 五、退款流程

`internal/modules/pay/service.go` Refund：

1. 校验：仅已支付(1)可退，退款金额 ≤ 支付金额
2. CAS `1→2`（退款中），失败说明已有并发退款在途
3. 调微信退款接口，带原单号 `out_trade_no` + 新退款单号（`TK` 前缀）
4. 受理失败 → 回滚 `2→1` 等待重试，绝不停留在中间态
5. 退款结果确认：真实模式目前停在「退款中」，**退款结果回调待接入**（service.go TODO）
6. 退款完成广播 `pay.refunded`，订单域消费后推进售后单

## 六、可靠性设计

| 风险 | 防线 |
|---|---|
| 回调重放 / 并发重复入账 | `MarkPaid` 乐观锁 `WHERE status=0`，`RowsAffected` 裁决胜负 |
| 前端谎报 / 掉单 | 前端无上报通道，入账只认回调 |
| 回调丢失 | 微信 24h 衰减重试；生产建议补主动查单兜底（待接入） |
| 退款停在中间态 | 受理失败回滚；结果回调待接入 |
| 金额被篡改 | 金额全部服务端计算（前端只传 id 与数量），调起参数 RSA 签名 |

## 七、代码索引

| 环节 | 位置 |
|---|---|
| 前端支付收口 | `web/apps/user-mp/src/utils/pay.ts` |
| 订单页发起支付 | `web/apps/user-mp/src/pages/order/detail.vue`（doPay） |
| 支付单创建 / 回调入账 | `server/internal/modules/pay/service.go` |
| 统一下单 + 签名 | `server/internal/infra/wxpay/jsapi.go` |
| 回调路由 | `server/internal/modules/pay/handler.go:26` |
| 乐观锁入账 | `server/internal/modules/pay/repo.go:44`（MarkPaid） |
| 事件消费注册 | `server/internal/bootstrap/bootstrap.go:76-79` |
| 状态常量 | `server/internal/modules/pay/model.go:13-16` |
| 配置（notify_url 等） | `server/configs/config.yaml` pay 段 |

## 八、上线检查清单

- [ ] `notify_url` 换成真实 HTTPS 域名（微信要求已备案域名），路径与后端路由一致，网关有前缀时注意改写
- [ ] 微信商户平台申请 API 证书，config 填 `mch_id`、商户私钥、序列号等
- [ ] 登录流程确保拿到并存储用户 `openid`（实名支付必需）
- [ ] 建议补充：超时关单 / 主动查单兜底、退款结果回调确认

## 附：本地联调

本地无公网收不到回调，开发配置走 mock 通道联调，`payments` 状态流转逻辑与真实模式完全一致。
