# task214-seepage · 多孔介质渗流试验边界反演服务

面向岩土/油气试验工程师的渗流边界反演后端服务：登记渗流柱试验，摄入入口/出口与
内部压力传感器序列，做一维达西瞬态正演与渗透率参数反演，判定边界结构可辨识性，
对比候选模型并发布反演结果。

## 业务域

多孔介质渗流试验中，试样两端的压力边界条件与内部渗透率分布需要从有限的传感器
观测中反演。本服务完成整条求解器生命周期：

1. 登记渗流柱试验（准备 → 采样 → 待反演 → 已封存）；
2. 幂等摄入入口/出口/内部压力序列（单位一致性 + 时间轴单调 + 缺口/异常检测）；
3. 登记边界候选模型（均匀 / 双层，可运行/可确认/已废弃状态机）；
4. 一维达西瞬态正演（隐式后向欧拉 + Thomas 三对角求解）；
5. 反演渗透率（均匀单参 / 双层交替坐标下降，含可辨识性区间）；
6. 对比候选模型判定结构可辨识性（稀疏观测不可辨识，加密观测可辨识）；
7. 发布反演结果（草稿 → 复核 → 发布 → 替代，不可变快照）。

## 标准命令

```bash
export GOTOOLCHAIN=local
CGO_ENABLED=0 go build ./...
CGO_ENABLED=0 go vet   ./...
CGO_ENABLED=0 go test  ./...
go run ./cmd/seepage --smoke-test   # 端到端自测，0 退出码即通过
go run ./cmd/seepage --addr :8080 --db ./seepage.db   # 启动 HTTP 服务
```

## API 入口

全部路由带 `/api` 前缀，共 26 个端点：试验 CRUD 与流转、传感器序列摄入/校验/统计、
边界模型登记/状态流转、反演执行/重跑/可辨识性、残差对比、发布与健康自检。

## 持久化

SQLite（`modernc.org/sqlite` v1.52.0，纯 Go 驱动，CGO 无关），5 表：
`experiments / sensor_sequences / boundary_models / inversion_tasks / release_versions`。
序列窗口指纹 UNIQUE 幂等，封存试验拒写，服务以文件 DB 启动时天然具备重启恢复。

## Docker 双架构

```bash
bash build_benzhi_docker.sh task214-seepage linux/amd64
bash build_benzhi_docker.sh task214-seepage linux/arm64
docker run --rm task214-seepage:latest --smoke-test
```
