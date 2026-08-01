// Node 25+ 在内置了全局 `localStorage` 绑定，但未传 --localstorage-file 时其值为
// undefined，导致 vitest 的 jsdom 环境不会覆盖该全局，任何直接访问 localStorage
// 的测试都会在 beforeEach/组件 setup 中拿到 undefined。这里在环境装配后把它
// 重新绑定到一个真实 jsdom 窗口的 Storage 上。
if (typeof globalThis.localStorage === 'undefined') {
  const { JSDOM } = await import('jsdom')
  const dom = new JSDOM('', { url: 'http://localhost/' })
  Object.defineProperty(globalThis, 'localStorage', {
    value: dom.window.localStorage,
    configurable: true,
    writable: true,
  })
}

export {}
