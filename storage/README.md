## storage

`storage` 目录承载各引擎的底层持久化实现，不直接处理 Redis 命令，也不负责引擎策略（如 flush 时机、compact 触发）。

### 子模块

- `bitcask/`：追加写数据文件 + 内存索引（key -> offset）
- `bptree/`：B+Tree 的 WAL 追加与回放
- `lsm/`：LSM 的 WAL 与 SSTable 文件读写

### 分层边界

- `engines/*`：负责“何时写、何时刷盘、何时 compact”等策略编排
- `storage/*`：负责“怎么编码、怎么落盘、怎么回放”的文件细节
