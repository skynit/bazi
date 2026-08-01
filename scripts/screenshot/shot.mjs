#!/usr/bin/env node
/**
 * Page screenshot helper for the bazi frontend.
 *
 * Usage:
 *   node scripts/screenshot/shot.mjs <name>:<route> [<name>:<route> ...]
 *
 * Examples:
 *   node scripts/screenshot/shot.mjs home:/ login:/login
 *   node scripts/screenshot/shot.mjs chart:/chart/{chartId} ziwei:/ziwei/{chartId}
 *
 * Behavior:
 * - Registers (or logs into) a dedicated screenshot account through the dev
 *   server API proxy and injects the JWT into localStorage before page load.
 * - Ensures at least one chart exists; `{chartId}` in a route is replaced
 *   with a real chart id.
 * - Captures a full-page screenshot at 1440x900 @2x into vue/screenshots/<name>.png
 *
 * Env overrides:
 *   SHOT_BASE_URL   dev server origin (default http://localhost:5174)
 */
import { mkdirSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
process.env.PLAYWRIGHT_BROWSERS_PATH ||= path.resolve(here, '../../vue/.pw-browsers')

// Dynamic import so PLAYWRIGHT_BROWSERS_PATH above is set before playwright loads.
const { chromium } = await import('playwright')

const BASE = (process.env.SHOT_BASE_URL || 'http://localhost:5174').replace(/\/$/, '')
const OUT_DIR = path.resolve(here, '../../vue/screenshots')
mkdirSync(OUT_DIR, { recursive: true })

const SHOT_USER = {
  username: 'shot_user',
  email: 'shot_user@example.com',
  password: 'shotpass1234',
}

async function api(pathname, { method = 'GET', token, body } = {}) {
  const res = await fetch(`${BASE}/api${pathname}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    ...(body ? { body: JSON.stringify(body) } : {}),
  })
  const text = await res.text()
  let json = null
  try {
    json = JSON.parse(text)
  } catch {
    // non-JSON response
  }
  return { status: res.status, json, text }
}

async function ensureToken() {
  let r = await api('/auth/register', { method: 'POST', body: SHOT_USER })
  if (r.status >= 400 || !r.json?.token) {
    r = await api('/auth/login', {
      method: 'POST',
      body: { username: SHOT_USER.username, password: SHOT_USER.password },
    })
  }
  if (!r.json?.token) throw new Error(`auth failed: HTTP ${r.status} ${r.text.slice(0, 200)}`)
  return r.json.token
}

async function ensureChartId(token) {
  const list = await api('/charts?page=1&page_size=1', { token })
  const items = list.json?.charts || list.json?.items || list.json?.list || []
  if (items.length > 0 && items[0].id) return items[0].id
  const payload = {
    birth_year: 1990,
    birth_month: 5,
    birth_day: 15,
    birth_hour: 10,
    birth_min: 30,
    birth_sec: 0,
    calendar_type: 'SOLAR',
    gender: 'MALE',
    zi_hour_policy: 'late_zi_next_day',
    name: '截图示例命盘',
    birth_place: '北京',
    timezone: 'Asia/Shanghai',
    use_true_solar_time: false,
  }
  const created = await api('/chart', { method: 'POST', token, body: payload })
  if (!created.json?.id) {
    throw new Error(`chart create failed: HTTP ${created.status} ${created.text.slice(0, 300)}`)
  }
  return created.json.id
}

const targets = process.argv.slice(2)
if (targets.length === 0) {
  console.error('usage: node shot.mjs <name>:<route> [<name>:<route> ...]')
  process.exit(1)
}

let token = null
let chartId = null
try {
  token = await ensureToken()
  chartId = await ensureChartId(token)
  console.log(`[shot] auth ok, chartId=${chartId}`)
} catch (e) {
  console.warn(`[shot] auth/chart setup failed (${e.message}); pages requiring auth will redirect to /login`)
}

const browser = await chromium.launch()
let failures = 0
for (const target of targets) {
  const idx = target.indexOf(':')
  if (idx <= 0) {
    console.error(`[shot] bad target "${target}", expected name:/route`)
    failures++
    continue
  }
  const name = target.slice(0, idx)
  const route = target.slice(idx + 1).replaceAll('{chartId}', String(chartId ?? 1))
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  })
  if (token) {
    await context.addInitScript((tok) => {
      try {
        localStorage.setItem('token', tok)
      } catch {
        // ignore
      }
    }, token)
  }
  const page = await context.newPage()
  try {
    await page.goto(`${BASE}${route}`, { waitUntil: 'load', timeout: 30000 })
    await page.waitForLoadState('networkidle', { timeout: 8000 }).catch(() => {})
    await page.waitForTimeout(1500)
    const file = path.join(OUT_DIR, `${name}.png`)
    await page.screenshot({ path: file, fullPage: true })
    console.log(`[shot] ${name} -> ${path.relative(process.cwd(), file)}`)
  } catch (e) {
    failures++
    console.error(`[shot] ${name} FAILED: ${e.message}`)
  }
  await context.close()
}
await browser.close()
if (failures > 0) process.exit(2)
