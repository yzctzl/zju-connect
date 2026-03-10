# ZJU Connect
## reconnect 分支主要改动

- **会话持久化与管理**
  - 支持会话保存至 `session.json`，重启无需重新扫码或短信验证。
  - **主动会话轮换 (12h)**：每 12 小时自动尝试更新会话，防止 24 小时硬过期，并支持失败后的平滑回退，确保连接不中断。
  - 即使会话文件标记为过期（>24h），仍会尝试与服务器“握手”恢复，最大限度复用旧会话。
- **智能分流与 DNS 优化**
  - 实现 DNS 智能分流机制，优先根据服务端下发的域名/IP 规则决定是否走 VPN，并在远端 DNS 不可用时平滑回退到本地解析。
  - 优化了直连切换及代理规则匹配。
- **重连策略与稳定性**
  - 引入分级重连策略（对会话过期和通用网络错误进行区分对待）。
  - 收发通道断开时优先重建数据通道，必要时才刷新会话或回退到完整登录，减少不必要的重新认证。
  - 修复了 `KeepAlive` 在某些边缘情况下的 panic 崩溃问题，极大提升了长期挂机的稳定性。
- **自动化建设**
  - 完善了 Docker 镜像及多平台二进制的构建 CI 流水线。


> 🚫 **免责声明**
>
> 本程序**按原样提供**，作者**不对程序的正确性或可靠性提供保证**，请使用者自行判断具体场景是否适合使用该程序，**使用该程序造成的问题或后果由使用者自行承担**！

---

中文 | [English](README_en.md)

**本程序基于 [EasierConnect](https://github.com/lyc8503/EasierConnect)（现已停止维护）完成，感谢原作者 [lyc8503](https://github.com/lyc8503)。**

**QQ 交流群：1037726410**，欢迎使用者加入交流。

### 使用方法

#### 使用 GUI 版客户端

+ 如果你是来自 ZJU 的用户：
  + Windows 用户推荐使用 [ZJU Connect for Windows](https://github.com/mythologyli/zju-connect-for-Windows)。
  + Linux/macOS 用户可以尝试使用 [PageChen04](https://github.com/PageChen04) 开发的客户端 [EZ4Connect](https://github.com/PageChen04/EZ4Connect) 或 [kowyo](https://github.com/kowyo) 开发的客户端 [hitsz-connect-verge](https://github.com/kowyo/hitsz-connect-verge)。
    注意请设置服务器地址为 `rvpn.zju.edu.cn:443`。
+ 如果你是非 ZJU 的用户：

  可以尝试使用 [PageChen04](https://github.com/PageChen04) 开发的客户端 [EZ4Connect](https://github.com/PageChen04/EZ4Connect) 或 [kowyo](https://github.com/kowyo) 开发的客户端 [hitsz-connect-verge](https://github.com/kowyo/hitsz-connect-verge)。

#### 直接运行

+ 如果你是来自 ZJU 的用户：

  1. 在 [Release](https://github.com/mythologyli/zju-connect/releases) 页面下载对应平台的最新版本。

  2. 以 macOS 为例，解压出可执行文件 `zju-connect`。

  3. macOS 需要先解除安全限制。命令行运行：`sudo xattr -rd com.apple.quarantine zju-connect`。

  4. 命令行运行：`./zju-connect -username <上网账户> -password <密码>`。

  5. 此时 `1080` 端口为 Socks5 代理，`1081` 端口为 HTTP 代理。如需更改默认端口，请参考参数说明。

+ 如果你是非 ZJU 的用户：

  其他步骤与上述相同，运行参数请尝试设置为：

  `./zju-connect -server <服务器地址> -port <服务器端口> -username xxx -password xxx -disable-zju-config -skip-domain-resource -zju-dns-server auto`

  *详情见此[链接](https://github.com/Mythologyli/zju-connect/issues/65#issuecomment-2650185322)*

#### 作为服务运行

[链接](docs/service.md)

#### Docker 运行

[链接](docs/docker.md)

### ⚠️ 警告

建议使用其他开启了 TUN 模式的代理工具时，同时配合 zju-connect 作为上游代理，因为实测中 zju-connect 可能会因为 direct 分流导致连接中断，本地客户端配置脚本[🔗](Script.js)。

### ⚠️ TUN 模式注意事项

1. 需要管理员权限运行

2. Windows 系统需要前往 [Wintun 官网](https://www.wintun.net)下载 `wintun.dll` 并放置于可执行文件同目录下

3. 为保证域名解析正确，建议配置 `dns-hijack` 劫持系统 DNS

### 连接与分流逻辑

1. 如果配置了 `session-file`，程序启动时会优先尝试恢复旧会话；只有当服务器确认旧会话不可用时，才会回退到完整登录流程。

2. 完整登录流程为：认证 -> 自动选线（默认开启）-> 申请 Token -> 拉取服务端资源规则 -> 申请虚拟 IP -> 建立收发数据通道。

3. 运行过程中如果 `send/recv` 数据通道异常，程序会先尝试重建数据通道；只有在判断更像是会话失效时，才会刷新 Token/IP 或重新登录。

4. 开启 ZJU DNS 时，命中 VPN 域名的查询会优先走远端 DNS；其他域名默认也会优先尝试远端 DNS，失败后再回退到备用/系统 DNS。最终是否走 VPN，仍由域名规则和 IP 规则共同决定。

5. TUN 模式的分流依赖系统路由、DNS 劫持和服务端下发的 IP 规则。为保证域名分流语义尽量正确，建议同时启用 `dns-hijack`，并根据需要开启 `add-route`。

### 参数说明

+ `server`: SSL VPN 服务端地址，默认为 `rvpn.zju.edu.cn`

+ `port`: SSL VPN 服务端端口，默认为 `443`

+ `username`: 网络账户。例如：学号

+ `password`: 网络账户密码

+ `totp-secret`: TOTP 密钥，可用于自动完成 TOTP 验证。如服务端无需 TOTP 验证或希望手动输入验证码，可不填

+ `cert-file`: p12 证书文件路径，如果服务器要求证书验证，需要配置此参数

+ `cert-password`: 证书密码

+ `disable-server-config`: 禁用服务端配置，一般不需要加此参数

+ `skip-domain-resource`: 不使用服务端下发的域名资源分流，一般不需要加此参数

+ `disable-zju-config`: 禁用 ZJU 相关配置，一般不需要加此参数

+ `disable-zju-dns`: 禁用 ZJU DNS 改用本地 DNS，一般不需要加此参数。关闭后只会禁用远端 DNS 探活，不再关闭网页 Session 保活和主动会话刷新；如果 `upstream-dns-mode` 使用 `remote-first` 或 `remote-only`，则必须保持 ZJU DNS 可用

+ `disable-multi-line`: 禁用自动根据延时选择线路。加此参数后，使用 `server` 参数指定的线路

+ `proxy-all`: 是否代理所有流量，一般不需要加此参数

+ `upstream-only`: 将 zju-connect 收敛为纯 EasyConnect 上游代理。开启后，HTTP/SOCKS/Shadowsocks 等入口的拨号都会强制走 VPN，并禁止内部 `direct` 兜底，适合在外层配合 `gost`、`mihomo` 等工具做分流

+ `upstream-dns-mode`: 上游模式下的 DNS 策略，支持 `auto`、`remote-first`、`remote-only`。默认是 `auto`；当 `upstream-only=true` 时，默认等效为 `remote-first`

+ `socks-bind`: SOCKS5 代理监听地址，默认为 `:1080`

+ `socks-user`: SOCKS5 代理用户名，不填则不需要认证

+ `socks-passwd`: SOCKS5 代理密码，不填则不需要认证

+ `http-bind`: HTTP 代理监听地址，默认为 `:1081`。为 `""` 时不启用 HTTP 代理

+ `shadowsocks-url`: Shadowsocks 服务端 URL。例如：`ss://aes-128-gcm:password@server:port`。格式[参考此处](https://github.com/shadowsocks/go-shadowsocks2)

+ `dial-direct-proxy`: 当 URL 未命中 RVPN 规则并切换到直连时使用代理，常用于与其他代理工具配合的场景，目前仅支持 http 代理。 例如：`http://127.0.0.1:7890"`，为 `""` 时不启用；`upstream-only=true` 时该参数会被忽略

+ `tun-mode`: TUN 模式（实验性）。请阅读后文中的 TUN 模式注意事项

+ `add-route`: 启用 TUN 模式时根据服务端下发的 IP 规则添加路由。不启用时，未禁用 ZJU 配置的情况下默认只添加 `10.0.0.0/8`

+ `dns-ttl`: DNS 缓存时间，默认为 `3600` 秒

+ `disable-keep-alive`: 禁用定时保活，一般不需要加此参数。关闭后会同时禁用远端 DNS 探活、网页 Session 保活和 12 小时主动会话刷新

+ `zju-dns-server`: ZJU DNS 服务器地址，默认为 `10.10.0.21`。设置为 auto 时使用从服务端获取的 DNS 服务器，如果未能获取则禁用 ZJU DNS

+ `secondary-dns-server`: 当使用 ZJU DNS 服务器无法解析时使用的备用 DNS 服务器，默认为 `114.114.114.114`。留空则使用系统默认 DNS，但在开启 `dns-hijack` 时必须设置

+ `dns-server-bind`: DNS 服务器监听地址，默认为空即禁用。例如，设置为 `127.0.0.1:53`，则可向 `127.0.0.1:53` 发起 DNS 请求

+ `dns-hijack`: 启用 TUN 模式时劫持 DNS 请求，建议在启用 TUN 模式时添加此参数。否则域名级分流规则在 TUN 模式下可能无法完整生效

+ `debug-dump`: 是否开启调试，一般不需要加此参数

+ `tcp-port-forwarding`: TCP 端口转发，格式为 `本地地址-远程地址,本地地址-远程地址,...`，例如 `127.0.0.1:9898-10.10.98.98:80,0.0.0.0:9899-10.10.98.98:80`。多个转发用 `,` 分隔

+ `udp-port-forwarding`: UDP 端口转发，格式为 `本地地址-远程地址,本地地址-远程地址,...`，例如 `127.0.0.1:53-10.10.0.21:53`。多个转发用 `,` 分隔

+ `custom-dns`: 指定自定义DNS解析结果，格式为 `域名:IP,域名:IP,...`，例如 `www.cc98.org:10.10.98.98,appservice.zju.edu.cn:10.203.8.198`。多个解析用 `,` 分隔

+ `custom-proxy-domain`: 指定自定义域名使用RVPN代理，格式为 `域名,域名,...`，例如 `nature.com,science.org`。多个域名用 `,` 分隔

+ `twf-id`: twfID 登录，调试用途，一般不需要加此参数

+ `session-file`: 指定会话保存文件路径（JSON 格式）。配置后可实现会话持久化，重启程序无需重新登录/短信验证。程序会优先尝试恢复旧会话，即便会话文件早于 24 小时也会尝试与服务器握手，恢复失败时才自动回退到完整登录流程

+ `keep-alive-domain`: 指定用于 Keep-Alive 检测的域名，默认为 `www.henu.edu.cn`

+ `config`: 指定配置文件，内容参考 `config.toml.example`。启用配置文件时其他参数无效

### 计划表

#### 已完成

- [x] 代理 TCP 流量
- [x] 代理 UDP 流量
- [x] SOCKS5 代理服务
- [x] HTTP 代理服务
- [x] Shadowsocks 代理服务
- [x] ZJU DNS 解析
- [x] ZJU 规则添加
- [x] 支持 IPv6 直连
- [x] DNS 缓存加速
- [x] 自动选择线路
- [x] TCP 端口转发功能
- [x] UDP 端口转发功能
- [x] 通过配置文件启动
- [x] 定时保活
- [x] TUN 模式
- [x] 自动劫持 DNS
- [x] 短信验证
- [x] TOTP 验证
- [x] 证书验证
- [x] 会话持久化与管理 (支持 `session.json` 保存与自动恢复)
- [x] 智能保活与断线重连优化 (分级重连策略)
- [x] 12 小时主动会话刷新 (带失败回退保护机制，确保长期挂机不掉线)
- [x] DNS 智能分流与解析策略优化

#### To Do

- [ ] TUN 模式下 `proxy-all` 的正确实现 (#64)

### 贡献者

<a href="https://github.com/mythologyli/zju-connect/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=mythologyli/zju-connect" />
</a>

### 感谢

+ [EasierConnect](https://github.com/lyc8503/EasierConnect)

+ [socks2http](https://github.com/zenhack/socks2http)

+ [![image](docs/yxvm.png)](https://yxvm.com/)

  [NodeSupport](https://github.com/NodeSeekDev/NodeSupport) 赞助了本项目
