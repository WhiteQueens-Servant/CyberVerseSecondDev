# Demo 数据写入执行说明

> 本文件供服务器端 Claude Code 或运维人员执行。将济南景区 demo 数据批量导入 `cyberverse.db`。

## 前置条件

- Go 后端已编译并成功启动过（确保 `data/cyberverse.db` 已创建，表结构已迁移）
- `sqlite3` 命令行工具可用

## 数据概览

| 表 | 记录数 | 说明 |
|---|---|---|
| `attractions` | 10 | 济南景点：趵突泉、大明湖、千佛山等 |
| `routes` | 4 | 推荐路线：泉水精华游、湖光山色深度游等 |
| `route_steps` | 14 | 路线途经点（每条路线 3-4 个景点） |

### 景点分类覆盖

| 分类 | 景点数 | 景点 |
|---|---|---|
| 历史古迹 | 3 | 趵突泉、千佛山、山东省博物馆 |
| 自然风光 | 4 | 大明湖、五龙潭、黑虎泉、红叶谷 |
| 亲子游乐 | 3 | 泉城广场、芙蓉街、济南野生动物世界 |

### 路线详情

| 路线 | 时长 | 难度 | 途经景点 |
|---|---|---|---|
| 泉水精华游 | 约2.5小时 | 简单 | 趵突泉→五龙潭→黑虎泉→泉城广场 |
| 湖光山色深度游 | 约5小时 | 适中 | 大明湖→芙蓉街→千佛山 |
| 亲子欢乐游 | 约6小时 | 简单 | 大明湖→芙蓉街→济南野生动物世界 |
| 文化研学游 | 约5小时 | 简单 | 山东省博物馆→趵突泉→大明湖 |

## 执行步骤

### 1. 备份现有数据（可选）

```bash
cp data/cyberverse.db data/cyberverse.db.bak
```

### 2. 执行 SQL 脚本

```bash
cd /path/to/CyberVerse-main
sqlite3 data/cyberverse.db < docs/second-development-plan/seed_jinan_demo.sql
```

### 3. 验证数据

```bash
# 检查景点数量
sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM attractions;"
# 预期：10

# 检查路线数量
sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM routes;"
# 预期：4

# 检查途经点数量
sqlite3 data/cyberverse.db "SELECT COUNT(*) FROM route_steps;"
# 预期：14

# 检查景点分类覆盖
sqlite3 data/cyberverse.db "SELECT category, COUNT(*) FROM attractions GROUP BY category;"
# 预期：
# 历史古迹|3
# 亲子游乐|3
# 自然风光|4

# 检查路线-景点关联
sqlite3 data/cyberverse.db "
SELECT r.name, a.name, rs.step_order, rs.duration_minutes, rs.highlight
FROM route_steps rs
JOIN routes r ON rs.route_id = r.id
JOIN attractions a ON rs.attraction_id = a.id
ORDER BY r.name, rs.step_order;"
# 预期：14 行，每条路线的途经景点名称正确显示

# 通过 API 验证
curl -s http://localhost:8080/api/v1/attractions | python3 -m json.tool | head -20
# 预期：返回 10 个景点的 JSON 数组

curl -s http://localhost:8080/api/v1/routes | python3 -m json.tool | head -20
# 预期：返回 4 条路线的 JSON 数组，steps 中 attraction_name 已正确填充
```

### 4. 重跑脚本（幂等性）

SQL 脚本使用 `INSERT INTO`（非 `INSERT OR REPLACE`），重复执行会因主键冲突报错。如需重跑：

```bash
# 清空后重跑
sqlite3 data/cyberverse.db "DELETE FROM route_steps; DELETE FROM routes; DELETE FROM attractions;"
sqlite3 data/cyberverse.db < docs/second-development-plan/seed_jinan_demo.sql
```

## 数据来源说明

- 景点信息基于济南公开旅游资源整理，描述为 demo 用途编写
- `image_url` 字段均为空（待后续补充景区实景图）
- 路线为虚构推荐，途经点的 `duration_minutes` 和 `highlight` 为估算值
- 所有 UUID 由 `crypto.randomUUID()` 生成，确保唯一性
