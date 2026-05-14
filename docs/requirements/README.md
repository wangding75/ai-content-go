# AI Content Factory 最新交付包（Go + 前后端整合版）

生成日期：2026-05-14

## 文件清单

| 文件 | 说明 |
|---|---|
| `docs/product/00-product-blueprint.md` | 最新项目蓝图，后端技术栈改为 Go |
| `docs/architecture/api-contract-standard.md` | API 契约规范 v5 |
| `docs/iterations/iteration-README.md` | 最新迭代总览 |
| `docs/iterations/iteration-*.md` | 每个迭代的产品、技术、前端页面、接口、映射、验收 |
| `docs/prototype-page-api-map.md` | 原型页面与接口总映射 |
| `prototype/ai-content-factory-clickable-prototype.html` | 完整可点击原型 |

## 关键调整

- 后端开发语言：Go / Golang。
- 前端不再单独成文档，已整合到各迭代。
- Iteration 2.1 保留为调度与接口契约补丁迭代。
- WorkflowSchedule / ProductionPlan 用于每天生成 5 个 ContentItem。
- n8n 只做外围自动化。
