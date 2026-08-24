基于 Go 实现的多孔介质渗流试验边界反演服务，一款纯后端工程分析服务，处理试验观测、边界参数反演与可追溯结果发布。

# BENZHI 评测说明 · task214-seepage

多孔介质渗流试验边界反演服务（纯 Go 后端，SQLite 持久化，无前端）。

## 运行命令

```bash
export GOTOOLCHAIN=local
CGO_ENABLED=0 go build ./...
CGO_ENABLED=0 go vet   ./...
CGO_ENABLED=0 go test  ./...
go run ./cmd/seepage --smoke-test
```

## HTTP API

监听 `:8080`，全部路由带 `/api` 前缀（26 个端点）：

- `POST /api/experiments` / `GET /api/experiments` / `GET /api/experiments/{id}`
- `PATCH /api/experiments/{id}/seal` / `PATCH /api/experiments/{id}/transition`
- `GET /api/experiments/{id}/geometry`
- `POST /api/experiments/{id}/sequences` / `GET /api/experiments/{id}/sequences`
- `GET /api/sequences/{id}` / `POST /api/sequences/{id}/revalidate`
- `GET /api/experiments/{id}/sequences/stats`
- `POST /api/experiments/{id}/models` / `GET /api/experiments/{id}/models`
- `GET /api/models/{id}` / `PATCH /api/models/{id}/status`
- `POST /api/models/{id}/inversions` / `GET /api/inversions/{id}`
- `GET /api/experiments/{id}/inversions` / `POST /api/inversions/{id}/rerun`
- `GET /api/experiments/{id}/residuals/compare`
- `GET /api/inversions/{id}/identifiability`
- `POST /api/experiments/{id}/releases` / `GET /api/experiments/{id}/releases`
- `GET /api/releases/{id}` / `GET /api/selfcheck`

## Docker 双架构

```bash
bash build_benzhi_docker.sh task214-seepage linux/amd64
bash build_benzhi_docker.sh task214-seepage linux/arm64
docker run --rm --platform linux/amd64 task214-seepage:latest --smoke-test
docker run --rm --platform linux/arm64 task214-seepage:latest --smoke-test
```

## --smoke-test 契约

`go run ./cmd/seepage --smoke-test`（及容器内 `/app/seepage --smoke-test`）不启动长驻
服务，而是用临时数据库跑完整端到端场景并以 0 退出码结束：

- 登记渗流柱试验并流转到待反演；
- 摄入入口压力边界序列 + 出口测量（稀疏观测）；
- 种下均匀/双层两个候选模型，对比判定为**不可辨识**（稀疏观测下两模型同残差）；
- 加密内部传感器（x=0.25/0.5/0.75），对比判定为**可辨识**（仅双层模型拟合）；
- 对双层模型执行反演并收敛（状态 converged），发布反演结果；
- 关闭并重开同一 SQLite，验证试验/序列/模型/反演/发布全部恢复。

任一断言失败返回退出码 1。
