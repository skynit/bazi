<script setup lang="ts">
defineProps<{
  placeholder?: boolean
}>()
</script>

<template>
  <section class="narrative-panel">
    <div class="np-kicker">月度总述</div>
    <p v-if="placeholder" class="np-text">
      本月概览按每日关系条数、五行样本与关键日期整理，可结合下方每日详记查看变化。
    </p>
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
  font-size: var(--fs-xs);
  color: var(--text-dim);
  margin-bottom: 0.45rem;
}

.np-text {
  margin: 0;
  color: var(--text-muted);
  line-height: 1.7;
  font-size: var(--fs-sm);
}
</style>
