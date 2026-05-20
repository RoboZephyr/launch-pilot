# Kit

gen-kit skill 在这里填充视觉 kit：

- `tokens.ts` — 品牌化 tokens（override globals.css 的基线 tokens）
- `components/` — 基础 UI 原件（Button / Card / Hero 等，复用 motion 库）

gen-page skill 的生成代码从这里 import 组件。

Template 初始状态为空，pipeline 首次跑 gen-kit 时填充。
