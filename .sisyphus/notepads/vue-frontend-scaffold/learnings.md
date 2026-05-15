## Vue Frontend Scaffolding — Learnings

### Files Created/Modified
- **vue/package.json** — Replaced. Added all deps: vue-router, pinia, axios, element-plus, echarts, vue-echarts, tailwindcss, @tailwindcss/vite
- **vue/vite.config.ts** — Replaced. Added tailwindcss() plugin alongside vue()
- **vue/src/style.css** — Replaced. Switched from manual CSS variables to `@import "tailwindcss"` + custom `@theme` tokens (bazi-red, bazi-blue, bazi-paper, bazi-ink)
- **vue/src/main.ts** — Replaced. Added Pinia, ElementPlus, router setup
- **vue/src/App.vue** — Replaced. Stripped HelloWorld, now just `<router-view />`
- **vue/src/router/index.ts** — New file. 6 routes with lazy-loaded views: /, /login, /register, /chart/:id, /fortune, /history
- **vue/src/views/*** — 6 new placeholder .vue files
- **vue/src/stores/.gitkeep** — New empty file
- **vue/src/api/.gitkeep** — New empty file
- **vue/Dockerfile** — New. Multi-stage: node:20-alpine build → nginx:alpine serve
- **vue/nginx.conf** — New. Proxies /api/ to backend:8080, serves SPA from /usr/share/nginx/html
- **nginx/default.conf** — New. Same content as vue/nginx.conf

### Files Preserved (not modified)
- tsconfig.json, tsconfig.app.json, tsconfig.node.json (per instructions)
- index.html (existing Vite scaffold)
- Existing assets (vite.svg, vue.svg, hero.png, favicon.svg, icons.svg)
- HelloWorld.vue component (left in place, not deleted)
- .gitignore, .vscode/, README.md

### Verification
- No LSP servers available (typescript-language-server, vue-language-server not installed — expected, no npm)
- All 29 files in vue/ confirmed present via glob
- nginx/default.conf confirmed present
