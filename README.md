# 双色球分析后端

参照 `box` 的 Go 项目结构初始化，使用 Go 1.25、Gin、Viper、GORM/MySQL 和 robfig/cron。

```text
backend/
  controller/        HTTP 接口
    dto/             请求和响应结构
  service/           分析业务逻辑
  models/            开奖数据库模型
config/              配置读取和数据库连接
cron/                定时任务调度扩展入口
middleware/          HTTP 中间件
pkg/response/        统一响应
router/              路由注册
utils/               号码校验
main.go              启动和优雅停机
```

## 本地启动

```sh
go mod tidy
go run .
```

默认监听 `http://localhost:8080`。当前 `config.yaml` 已启用数据库及开奖表自动迁移，请确认 MySQL 连接配置后启动。检查和构建：

```sh
go test ./...
go vet ./...
go build -o bin/ssq-server .
```

## 接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/health` | 服务存活检查 |
| POST | `/api/v1/analysis` | 单期号码校验与描述统计 |
| GET | `/api/v1/records?limit=200&offset=0` | 分页查询开奖历史，limit 最大 500 |

请求示例（期号和日期无需传入）：

```json
{"red_balls": [1, 8, 12, 19, 25, 33], "blue_ball": 6}
```

成功响应：

```json
{
  "code": 200,
  "msg": "分析成功",
  "data": {
    "red_balls": [1, 8, 12, 19, 25, 33],
    "blue_ball": 6,
    "sum": 98,
    "span": 32,
    "odd_count": 4,
    "even_count": 2,
    "zone_counts": [2, 2, 2]
  }
}
```

红球必须为 6 个不重复的 1–33 整数，蓝球必须为 1–16 整数。输出红球升序排列；和值、跨度、奇偶个数均按红球计算。三区依次为 1–11、12–22、23–33。非法输入返回 HTTP 400。

## 配置与数据库

默认读取根目录 `config.yaml`。可通过 `SSQ_CONFIG` 指定另一个配置文件，所有字段支持 `SSQ_` 前缀环境变量覆盖，如 `SSQ_SERVER_PORT`、`SSQ_DATABASE_ENABLED`、`SSQ_DATABASE_DSN`。

开奖数据源通过 `source_url` 配置，定时抓取和在线历史导入共用此地址，也可用环境变量 `SSQ_SOURCE_URL` 覆盖。地址必须使用 HTTP 或 HTTPS；修改后重启后端（Docker 部署执行 `docker compose restart shuangseqiu`）。

```yaml
source_url: http://data.17500.cn/ssq_asc.txt
```

要启用 MySQL，先创建数据库，再设置：

```yaml
database:
  enabled: true
  driver: mysql
  dsn: "user:password@tcp(127.0.0.1:3306)/shuangseqiu?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: true
```

建议通过环境变量提供实际 DSN。`auto_migrate` 开启时创建或更新 `ssq_records` 表；表模型以 `issue` 为主键，没有额外 ID、创建时间或更新时间字段。也可手动执行 `deploy/sql/ssq_records.sql`，然后关闭自动迁移。

## 开奖同步任务

后端启动并连接数据库后注册任务，每周一、三、五东八区上午 06:00 执行（秒级表达式 `0 0 6 * * 1,3,5`）。使用固定 UTC+8 时区，不受主机或容器时区影响。

任务通过 GET 请求 [17500 双色球数据文件](http://data.17500.cn/ssq_asc.txt)，筛选东八区前一自然日的开奖结果。前 29 列依次对应期号、日期、6 个排序红球、蓝球、6 个出球顺序红球、投注总额、奖池金额，以及一至六等奖注数/金额。31 列格式末尾的两列附加数据忽略。

号码、出球顺序、整数金额及注数通过校验后才入库。使用 MySQL `ON DUPLICATE KEY UPDATE` 按期号更新，重复执行不会产生重复记录；源文件尚未提供目标日期、返回 HTTP 错误或数据不合法时记录失败日志。每次请求超时 30 秒，整个任务限时 2 分钟。

后端进程需要持续运行。任务不会在启动时立即触发，也不补跑停机期间错过的任务。数据库关闭时不会注册此任务。

## 福彩 3D

前端通过顶部玩法切换进入 `/#/fucai3d`，双色球入口为 `/#/ssq`。3D 按百、十、个位统计 0–9 的频次，允许重复数字和 0；保留位置顺序。沿用 100～1000 期十档样本、按实际样本归一化的频次色阶、历史对照、号码选择及随机组合。组合分析包含和值、跨度、奇偶比例、大小比例、豹子/组三/组六形态。历史对照不将目标期计入样本。

`config.yaml` 中的 `fucai3d_source_url` 配置 3D 数据源，默认 `http://data.17500.cn/3d_asc.txt`，环境变量 `SSQ_FUCAI3D_SOURCE_URL` 可覆盖。独立表 `fucai3d_records` 对应数据文件 17 列；SQL 位于 `deploy/sql/fucai3d_records.sql`。后端启动时根据 `database.auto_migrate` 自动建表。

3D 定时任务每天东八区 06:00 同步前一日开奖，表达式 `0 0 6 * * *`，重复同步按期号更新。每次下载超时 30 秒，任务限时 2 分钟。休市或尚未发布的日期只记录日志，不补造记录。双色球任务保持每周一、三、五 06:00。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/api/v1/fucai3d/records?limit=500&offset=0` | 3D 开奖历史分页 |
| POST | `/api/v1/fucai3d/analysis` | 3D 组合分析，请求 `{"numbers":[0,5,0]}` |

导入 3D 历史（全文件校验后事务写入，并核验期号数量）：

```sh
go run ./cmd/import-history -game fucai3d
go run ./cmd/import-history -game fucai3d -file .cache/3d_asc.txt -validate-only
docker compose exec shuangseqiu ./import-history -game fucai3d
```

## 手动导入双色球历史数据

```sh
go run ./cmd/import-history
```

下载并校验全部历史数据后，在同一数据库事务中按每批 200 期更新或插入。失败时整体回滚；成功后查询数据库核验导入期号数量，并输出表内总期数、期号与日期范围。重复执行不会新增重复期号。

仅下载校验、不连接数据库：`go run ./cmd/import-history -validate-only`。
也可通过 `-file 数据文件路径` 从本地文件导入。
通过 `-save-source .cache/ssq_asc.txt` 保存在线下载校验后的原始历史文件，便于后续离线导入。

## Docker

参考 `homebox`，使用 Docker Engine / Docker Desktop 和 Docker Compose v2 部署单容器。Supervisor 在同一个容器内运行 Nginx 和 Go 后端，并在子进程异常退出时自动重启。项目根目录 `config.yaml` 只读挂载到 `/app/config.yaml`，镜像不包含其中的数据库账号；连接现有外部 MySQL。

部署前确认 `config.yaml` 中数据库地址可从容器访问，且 `database.enabled: true`。现有历史记录无需重新导入。容器中的 `127.0.0.1` 指容器自身；远程 MySQL 使用实际 IP 或域名。Docker Desktop 上的宿主机 MySQL 可使用 `host.docker.internal:3306`；Linux 上若要使用此名称，在 Compose 服务下添加 `extra_hosts: ["host.docker.internal:host-gateway"]`。

```sh
docker compose up --build -d --remove-orphans
docker compose ps
docker compose logs -f shuangseqiu
```

访问 `http://localhost:8088`。Nginx 监听容器 80 端口，将 `/api` 和 `/health` 转发到同容器的 `127.0.0.1:8080`。镜像通过 `SSQ_SERVER_PORT=8080` 和 `SSQ_SERVER_MODE=release` 固定内部后端端口及模式。健康检查同时确认首页与后端响应；容器持续运行时，双色球和福彩3D的东八区定时同步继续生效。

修改访问端口：调整 `docker-compose.yml` 中 `ports` 左侧的 `8088`。若从旧双容器版本升级，`--remove-orphans` 会移除旧的 `backend`、`frontend` 容器。

容器内手动导入历史数据：

```sh
docker compose exec shuangseqiu ./import-history
```

更新代码后执行 `docker compose up --build -d`。停止服务使用 `docker compose down`，外部 MySQL 数据不会随容器停止而删除。

Dockerfile 分别构建前端与后端，再打包进同一个运行镜像。也可以直接构建：

```sh
docker build -t shuangseqiu:local .
```

## H5 前端

`cfront` 为 Vue 3 + Vite H5 项目，包含出球位置频次矩阵、号码选择、历史对照、组合特征分析和开奖历史浏览。先启动 Go 后端，再进入 `cfront` 执行 `npm install` 与 `npm run dev`，访问 `http://127.0.0.1:5173`。详情见 `cfront/README.md`。

当前提供单期描述统计、历史数据手动导入、分页开奖查询和前一日开奖定时入库；H5 按真实开奖记录计算位置频次，跨期遗漏等其他分析可继续扩展。
