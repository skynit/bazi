<script setup lang="ts">
defineProps<{
  placeholder?: boolean
}>()
</script>

<template>
  <section class="narrative-panel">
    <div class="np-kicker">月度总述</div>
    <p v-if="placeholder" class="np-text">本月走势已按每日评分、五行分布与关键日期生成，可结合下方每日详记查看节奏变化。</p>
    <slot v-else />
  </section>
</template>

<style scoped>
.narrative-panel {
  border: 1px solid var(--glass-border);
  border-radius: 0.75rem;
  padding: 1rem;
  background: var(--glass);
  color: var(--text);
}

.np-kicker {
  font-size: 0.76rem;
  color: var(--text-dim);
  margin-bottom: 0.45rem;
}

.np-text {
  margin: 0;
  color: var(--text-muted);
  line-height: 1.7;
  font-size: 0.9rem;
}
</style>
