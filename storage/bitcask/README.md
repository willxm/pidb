## storage/bitcask

Bitcask 存储层实现：维护数据文件追加写与按 offset 读取能力，由 `engines/bitcask.go` 的引擎编排调用。

### 文件职责

- `engine.go`：Bitcask 存储流程（索引构建、Set/Get/Del）
- `data_file.go`：数据文件抽象（扫描、追加、按偏移读取）
- `record.go`：记录编解码（文本行协议）

### 数据格式

当前行格式（CSV-like）：

- 写入：`key,value\n`
- 删除：`key,\n`

启动时扫描全文件重建索引：

- 非 tombstone 记录：`index[key] = offset`
- tombstone 记录：删除 `index[key]`

### 说明

- 读取路径通过内存索引定位 offset，再回读单条记录
- 当前实现偏教学/实验风格，未做 segment 拆分与 merge 压缩
