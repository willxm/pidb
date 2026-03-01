## storage/bptree

B+Tree 持久化日志模块（WAL），用于记录变更并在重启时回放恢复。

### 文件职责

- `wal.go`：WAL 打开、追加、回放、关闭

### 日志格式

按行存储，字段以 `\t` 分隔，key/value 使用 Base64 编码：

- `S\t<base64(key)>\t<base64(value)>`：设置 key
- `D\t<base64(key)>`：删除 key

### 使用边界

- 仅负责日志 I/O 与解析
- 不负责 B+Tree 插入/删除算法本身（由 `engines/bp_tree.go` 负责）
