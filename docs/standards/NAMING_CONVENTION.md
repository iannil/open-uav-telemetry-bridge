# 文档命名规范

## 文件命名格式

### 进度文档 (`/docs/progress/`)
```
PROGRESS_[功能模块]_[YYYYMMDD].md
```
示例：
- `PROGRESS_MAVLINK_ADAPTER_20260118.md`
- `PROGRESS_CORE_KERNEL_20260120.md`

### 完成报告 (`/docs/reports/completed/`)
```
COMPLETED_[功能模块]_[版本号]_[YYYYMMDD].md
```
示例：
- `COMPLETED_MAVLINK_ADAPTER_v0.1_20260125.md`
- `COMPLETED_MQTT_PUBLISHER_v0.1_20260130.md`

### 验收报告 (`/docs/reports/`)
```
REVIEW_[功能模块]_[版本号]_[YYYYMMDD].md
```
示例：
- `REVIEW_MAVLINK_ADAPTER_v0.1_20260126.md`

### 实施计划
```
IMPLEMENTATION_PLAN.md          # 主计划文档
IMPLEMENTATION_PLAN_[模块].md   # 模块级计划
```

## 文档内容规范

### 必须包含的元信息
- 创建时间
- 最后更新时间
- 作者/修改者
- 文档状态（进行中/已完成/已归档）

### 进度文档必须包含
- 当前进展摘要
- 已完成事项
- 进行中事项
- 待处理事项
- 遇到的问题与解决方案
- 下一步计划

### 完成报告必须包含
- 功能概述
- 实现细节
- 测试结果
- 已知限制
- 相关文件列表

## 版本号规范

遵循语义化版本：`v主版本.次版本.修订号`
- v0.1.0 - MVP 版本
- v0.2.0 - 新功能版本
- v0.1.1 - Bug 修复版本
