<script setup lang="ts">
import { ref } from 'vue'

const dialogVisible = ref(false)

function openDialog() {
  dialogVisible.value = true
}
</script>

<template>
  <div class="flex justify-center py-4">
    <!-- Grey/disabled-style button -->
    <button
      class="inline-flex cursor-pointer items-center gap-2 rounded-md border border-gray-300 bg-gray-100 px-5 py-2.5 text-sm font-medium text-gray-500 shadow-sm transition-all hover:bg-gray-200 hover:text-gray-600"
      @click="openDialog"
    >
      <span class="text-lg">🤖</span>
      <span>AI 智能分析</span>
    </button>
  </div>

  <!-- El Dialog modal -->
  <el-dialog
    v-model="dialogVisible"
    title="AI 智能分析"
    width="420px"
    align-center
  >
    <div class="flex flex-col items-center gap-4 py-6">
      <!-- Decorative icon -->
      <div class="flex h-20 w-20 items-center justify-center rounded-full bg-[var(--color-bazi-paper)]">
        <span class="text-4xl">🤖</span>
      </div>
      <p class="text-center text-base text-[var(--color-bazi-ink)] leading-relaxed">
        AI分析功能即将上线，敬请期待
      </p>
      <p class="text-center text-sm text-gray-400">
        我们将引入智能命理分析，为您提供更深入的解读
      </p>
    </div>
    <template #footer>
      <el-button type="primary" @click="dialogVisible = false">我知道了</el-button>
    </template>
  </el-dialog>
</template>
