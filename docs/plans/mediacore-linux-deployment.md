# Mediacore Linux 部署

一期沿用上游 `docker-compose.deploy.yml` 的 PostgreSQL、Redis、迁移、后端、Web 和独立 Agent。源码构建叠加 `docker-compose.build.yml`，后端补装原生媒体处理需要的 FFmpeg/ffprobe，不新增业务模块。

## 当前边界

首次部署按用户要求停用模型。用户随后确认先供内部试用：启用首批文本、图片，每次 1 积分，保留已有视频设置；再补齐示范文件的全部九型号，新增六项每次 1 积分。配置见[原生渠道说明](mediacore-4aion-native-config.md)。此次试用价格有单独授权，不是正式商业售价；商业售价仍待确认。不新增付费测试，由用户在画布试用。公开注册在 SMTP 就绪前关闭，保留原生邮箱验证。

## 运行目录与版本

本工程独立目录为 `/srv/mediacore-yingce`，Compose 项目为 `mediacore-yingce`。其他工程的容器、目录和数据保留。

- `releases/<完整提交 SHA>`：提交的源码及原生 Compose。
- `current`：当前运行版本的链接。
- `.env`：数据库密码、Agent Token、镜像引用和运行配置，仅 root 可读。
- `state/backend`：原生素材及 `.settings-key`；`state/postgres`、`state/redis`、`state/web`：各服务的持久目录。
- `ops`：本次构建记录、镜像内容 ID、私密初始化资料及备份脚本。
- `backups`：数据库、素材、解密设置所需文件和部署配置的常规备份。

首次部署的三个自有镜像由提交 `26cbbe5eb69100da7ed37dabcd6e7223086fbfc0` 构建，以完整 SHA 为独立标签，并记录 Docker 内容 ID。Web 与后端通过 `BUILD_COMMIT`、`BUILD_TIME` 显示构建来源；Agent 使用同一份源码。这是自有镜像部署，不宣称已发布 GHCR digest 版本。上游 Host Updater 未安装，不使用它覆盖本次源码部署。

### Agent 审批 Web 修复

修复提交 `f6253020a169abc0ad8e2463fd8d9562510ffc22` 仅调整前端审批发送字段并新增回归检查，后端和 Agent 源码相对初次部署逐字节未变，因此只更新 Web，不重启后端或 Agent，也不替用户确认待审批任务。运行记录按服务保留源码版本，不能将后端仍显示初次部署提交当作本次 Web 未更新。

服务器全量 Docker 构建先遇到 Docker Hub 鉴权连接超时，使用代理后仍阻塞在依赖下载；该构建已停止，未当作通过。实际 Web 产物来自同一提交的本机 Bun 1.3.9/Vite 生产构建，无本机 API 地址或 `.env` 覆盖。2054 个产物文件在服务器逐一校验 SHA-256 后，复用原有 Nginx 镜像及原生 `CANVAS_WEB_RELEASE_DIR` 发布入口构建运行镜像，标签为完整修复 SHA。没有改生产 Dockerfile、锁文件或静态发布脚本，也没有绕过依赖完整性检查。

保留旧 Web 镜像、原 `.env`、运行目录链接及镜像记录以供回滚；镜像来源与逐服务版本记在 `ops/images.json` 和 `ops/web-f625302-runtime.json`。Web 回滚不恢复业务数据，用户审批仍需手动确认。

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
- 首次部署时三个 4aion 渠道和模型均停用，售价未配置，普通目录为空。后续按用户确认开放内部试用，当前目录为文本、图片、视频各三型号；新增六项的能力筛选和原生报价每次 1 积分通过，已有三项配置保留。个人渠道、字幕转写、RunningHub 和整个素材转换插件继续用原生开关关闭。不配置独立音频渠道或方舟 IAM。
- 原生 Worker/全局渠道并发配置为 3，每用户活动任务上限为 1。后者同时包含排队与运行中的所有任务，直接沿用原生策略。10 人编辑仍是规模参考，未进行专项压测。
- 正式 HTTPS 首页、登录页和注册关闭提示已在浏览器查看，品牌为 Mediacore。原域名代理配置保存在 `backups/canvas-before.conf`；`nginx -t` 通过并已重载，其他域名配置未改。外层代理沿用上游 SSE 的 `X-Accel-Buffering: no` 信号。

这些检查不产生 AI 调用。模型可选与原生报价不等于完整生成闭环；线上生成、结果保存/下载及画布读回由用户试用确认。SMTP、公开注册和商业售价继续待配置。此前本机三个授权模型的实测结果见原生渠道说明，未重复进行全套上游验收。

## 备份

部署完成后的备份为 `backups/20261013T153404Z`，包含 PostgreSQL 自定义格式 dump、后台素材及 `.settings-key`、私密部署配置和镜像记录。已确认 dump 可列出内容、素材归档包含解密设置所需的 key；没有宣称完成恢复演练。此前首次备份也保留。

`/etc/cron.d/mediacore-yingce` 每日服务器本地时间 03:30 运行以下脚本，日志在 `ops/backup.log`。当前备份保留在本机，不自动删除历史备份。

```sh
/srv/mediacore-yingce/ops/mediacore-yingce-backup.sh
```

恢复需同时处理数据库、后台数据目录及部署配置，不能只恢复代理或只导入数据库。
