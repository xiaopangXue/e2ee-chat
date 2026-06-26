# E2EE Chat Deploy Package

解压后目录内直接运行：

```bash
chmod +x ./e2ee-chat ./start.sh
./start.sh
```

`start.sh` 内置默认环境变量：

```bash
ADDR=0.0.0.0:8080
TRUSTED_PROXIES=cloudflare
POW_DIFFICULTY=12
```

需要修改时可以直接编辑 `start.sh`，也可以启动前用环境变量覆盖：

```bash
ADDR=127.0.0.1:8080 POW_DIFFICULTY=16 ./start.sh
```

然后在 Cloudflare/Nginx 中反代到该服务。

Windows 测试可运行：

```powershell
.\bin\e2ee-chat-windows-amd64.exe
```

注意：

- 服务端不保存消息和文件。
- `static/` 必须和可执行文件在同一目录。
- 群聊码模式不适合敏感内容，尤其是短自定义码；完整邀请链接模式更安全。
