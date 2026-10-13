# Mediacore Linux 部署

一期沿用上游 `docker-compose.deploy.yml` 的 PostgreSQL、Redis、迁移、后端、Web 和独立 Agent。源码构建叠加 `docker-compose.build.yml`，后端补装原生媒体处理需要的 FFmpeg/ffprobe，不新增业务模块。

## 当前边界

用户选择先部署，模型保持停用，不新增付费测试。三个 4aion 渠道的配置见[原生渠道说明](mediacore-4aion-native-config.md)，生产不复制本机每次 1 积分的测试价格。SMTP 和商业积分售价待提供；公开注册在 SMTP 就绪前关闭，保留原生邮箱验证。

## 运行目录与版本

本工程独立目录为 `/srv/mediacore-yingce`，Compose 项目为 `mediacore-yingce`。其他工程的容器、目录和数据保留。

- `releases/<完整提交 SHA>`：提交的源码及原生 Compose。
- `current`：当前运行版本的链接。
- `.env`：数据库密码、Agent Token、镜像引用和运行配置，仅 root 可读。
- `state/backend`：原生素材及 `.settings-key`；`state/postgres`、`state/redis`、`state/web`：各服务的持久目录。
- `ops`：本次构建记录、镜像内容 ID、私密初始化资料及备份脚本。
- `backups`：数据库、素材、解密设置所需文件和部署配置的常规备份。

三个自有镜像均由提交 `26cbbe5eb69100da7ed37dabcd6e7223086fbfc0` 构建，以完整 SHA 为独立标签，并记录 Docker 内容 ID。Web 与后端通过 `BUILD_COMMIT`、`BUILD_TIME` 显示构建来源；Agent 使用同一份源码。这是服务器源码部署，不宣称已发布 GHCR digest 版本。上游 Host Updater 未安装，不使用它覆盖本次源码部署。

构建使用 `ALPINE_MIRROR=https://mirrors.aliyun.com/alpine` 和原有 `GOPROXY` 参数选择包下载源；默认配置仍保持上游源。Docker 镜像下载使用服务器已有代理，未改其他工程的代理或 Docker 配置。

前端第一次下载部分依赖触发完整性校验失败。使用构建进程的代理重新获取后，原生 Bun 1.3.9 `--frozen-lockfile` 安装和生产镜像构建通过，未修改锁文件或跳过校验。本机同源 Vite 构建亦通过；实际运行的是服务器构建的镜像。临时构建器仅属于本工程，完成后移除，不修改其他构建器。

## 基本操作

在服务器 root 会话中：

```sh
cd /srv/mediacore-yingce/current
docker compose --project-name mediacore-yingce --env-file ../../.env -f docker-compose.deploy.yml ps
```

Web 仅绑定本机 `127.0.0.1:13000`，由现有 Nginx 和 HTTPS 证书提供 `https://canvas.xingcyj.com`。PostgreSQL、Redis、后端和 Agent 不发布宿主机端口。切换前保存该域名的原代理配置，恢复代理不会恢复业务数据。

## 已完成检查

- 五个常驻服务均健康，独立迁移服务退出码为 0；FFmpeg/ffprobe 在后端容器内执行成功。Web 静态包和后端健康接口均包含本次构建提交。
- 通过原生接口完成受控管理员初始化、普通账号登录、创建并保存画布、兑换 1 积分及账本检查、PNG 上传与鉴权下载。重启本工程后端后仍可读取画布与素材；切换域名后又通过 HTTPS 登录和读取同一数据。检查账号最后停用，记录保留，不计入运营账号。
- 三个 4aion 渠道和模型均停用，售价未配置；普通目录为空且不含供应商密钥或地址。个人渠道、字幕转写、RunningHub 和整个素材转换插件均用原生开关关闭。不配置独立音频渠道或方舟 IAM。
- 原生 Worker/全局渠道并发配置为 3，每用户活动任务上限为 1。后者同时包含排队与运行中的所有任务，直接沿用原生策略。10 人编辑仍是规模参考，未进行专项压测。
- 正式 HTTPS 首页、登录页和注册关闭提示已在浏览器查看，品牌为 Mediacore。原域名代理配置保存在 `backups/canvas-before.conf`；`nginx -t` 通过并已重载，其他域名配置未改。外层代理沿用上游 SSE 的 `X-Accel-Buffering: no` 信号。

这些检查不产生 AI 调用。服务器真实生成、SMTP、公开注册、商业售价和启用模型继续待配置；此前本机三个授权模型的实测结果见原生渠道说明。用户试用尚待确认，未重复进行全套上游验收。

## 备份

部署完成后的备份为 `backups/20261013T153404Z`，包含 PostgreSQL 自定义格式 dump、后台素材及 `.settings-key`、私密部署配置和镜像记录。已确认 dump 可列出内容、素材归档包含解密设置所需的 key；没有宣称完成恢复演练。此前首次备份也保留。

`/etc/cron.d/mediacore-yingce` 每日服务器本地时间 03:30 运行以下脚本，日志在 `ops/backup.log`。当前备份保留在本机，不自动删除历史备份。

```sh
/srv/mediacore-yingce/ops/mediacore-yingce-backup.sh
```

恢复需同时处理数据库、后台数据目录及部署配置，不能只恢复代理或只导入数据库。
