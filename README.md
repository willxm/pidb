## pidb

一个基于 Redis 协议（`redcon`）的轻量 KV 数据库实验项目。

### 目录结构

```text
.
├── cmd/
│   └── main.go                # 服务入口，命令解析与网络层
├── engines/                   # 逻辑引擎层（实现统一 Engine 接口）
│   ├── mem.go
│   ├── bp_tree.go
│   ├── lsm_tree.go
│   └── bitcask.go
├── storage/                   # 具体存储实现
│   ├── bitcask/
│   │   ├── engine.go
│   │   ├── data_file.go
│   │   ├── record.go
│   │   ├── record_test.go
│   │   └── engine_test.go
│   ├── bptree/
│   │   └── wal.go
│   └── lsm/
│       ├── record.go
│       ├── wal.go
│       └── sstable.go
├── test/
│   └── redis.go               # go-redis 客户端联调脚本
├── engine.go                  # Engine 接口定义
├── mem.go                     # 向后兼容：类型别名导出
├── bp_tree.go                 # 向后兼容：类型别名导出
├── lsm_tree.go                # 向后兼容：类型别名导出
├── bitcask.go                 # 向后兼容：类型别名导出
└── engine_integration_test.go # 跨引擎一致性集成测试
```

### 分层说明

- `cmd`：网络协议和命令分发，不关心具体存储细节。
- `engine.go`：定义统一能力（`New/Set/Get/Del/Exists`）。
- `engines`：面向业务的引擎抽象与组合。
- `storage`：底层持久化细节（bitcask datafile/record、bptree WAL、lsm wal/sstable）。