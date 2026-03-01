## storage/lsm

LSM 存储层模块，提供 WAL 与 SSTable 的文件读写能力，供 `engines/lsm_tree.go` 调用。

### 文件职责

- `record.go`：LSM 通用记录结构（值与 tombstone）
- `wal.go`：WAL 追加、回放、重置
- `sstable.go`：SSTable 列举、写入、单 key 查询、全量读取

### WAL 格式

按行存储，字段 `\t` 分隔，key/value Base64 编码：

- `S\t<key>\t<value>`
- `D\t<key>`

### SSTable 格式

同样使用逐行文本记录：

- `S\t<key>\t<value>`：有效值
- `D\t<key>`：删除标记（tombstone）

### 设计意图

- 引擎层决定 flush / compact 策略
- 存储层只提供“可复用的文件操作原语”
