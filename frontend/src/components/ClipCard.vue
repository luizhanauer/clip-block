<script setup lang="ts">
import { format } from "date-fns";
import {
  Trash2,
  Pin,
  Check,
  Code2,
  Wand2,
  Braces,
  Minimize,
  RotateCcw,
  ArrowUpAZ,
  ArrowDownAZ,
} from "lucide-vue-next";
import { ref, onMounted, computed } from "vue";
import hljs from "highlight.js/lib/core";
import javascript from "highlight.js/lib/languages/javascript";
import go from "highlight.js/lib/languages/go";
import json from "highlight.js/lib/languages/json";
import bash from "highlight.js/lib/languages/bash";
import "highlight.js/styles/atom-one-dark.css";

import type { Clip } from "../types/Clip";
import {
  formatJSON,
  minify,
  toUpperCase,
  toLowerCase,
} from "../utils/textTools";

hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("go", go);
hljs.registerLanguage("json", json);
hljs.registerLanguage("bash", bash);

const props = defineProps<{
  clip: Clip;
  selected?: boolean;
}>();

const emit = defineEmits(["delete", "pin", "copy", "toggle-select"]);

const wasCopied = ref(false);
const showTools = ref(false);
const isCode = ref(false);
const highlightedContent = ref("");
const displayContent = ref(props.clip.content);

const dateStr = format(new Date(props.clip.created_at), "dd/MM HH:mm:ss");

const isModified = computed(() => displayContent.value !== props.clip.content);

const detectCode = (text: string) => {
  if (
    /^(\{|\[|func|import|const|let|var|<\?php|package|class)/.test(
      text.trim(),
    ) ||
    text.includes(";")
  ) {
    try {
      const result = hljs.highlightAuto(text);
      highlightedContent.value = result.value;
      isCode.value = true;
    } catch {
      isCode.value = false;
    }
  } else {
    isCode.value = false;
  }
};

onMounted(() => detectCode(displayContent.value));

const resetContent = () => {
  displayContent.value = props.clip.content;
  detectCode(displayContent.value);
};

const applyTool = (tool: "json" | "mini" | "upper" | "lower") => {
  let newText = displayContent.value;
  switch (tool) {
    case "upper":
      newText = toUpperCase(newText);
      break;
    case "lower":
      newText = toLowerCase(newText);
      break;
    case "json":
      newText = formatJSON(newText);
      break;
    case "mini":
      newText = minify(newText);
      break;
  }
  displayContent.value = newText;
  detectCode(newText);
};

const handleClick = (event: MouseEvent) => {
  if (event.ctrlKey || event.shiftKey) emit("toggle-select", props.clip.id);
  else handleCopy();
};

const handleCopy = () => {
  emit("copy", displayContent.value);
  wasCopied.value = true;
  setTimeout(() => (wasCopied.value = false), 1000);
};
</script>

<template>
  <div
    class="relative group w-full mb-3 rounded-lg border-l-4 transition-all duration-200 cursor-pointer shadow-sm bg-surface border-y border-r hover:translate-x-1"
    :class="[
      selected
        ? 'border-accent ring-1 ring-accent bg-overlay'
        : clip.is_pinned
          ? 'border-l-warn border-overlay'
          : 'border-l-transparent border-overlay',
      selected ? '' : 'hover:bg-overlay hover:shadow-md',
    ]"
    @click="handleClick"
    @mouseleave="showTools = false"
  >
    <div class="p-3 pr-8">
      <div
        v-if="isCode"
        class="font-mono text-xs overflow-x-auto bg-[#11111b] p-2 rounded border border-white/5"
      >
        <pre><code v-html="highlightedContent"></code></pre>
      </div>
      <div
        v-else
        class="font-mono text-sm text-text break-all line-clamp-3 leading-relaxed whitespace-pre-wrap"
      >
        {{ displayContent }}
      </div>

      <div class="flex justify-between items-end mt-2 h-4">
        <div class="flex items-center gap-2">
          <span
            class="text-[10px] text-subtext font-bold bg-bg px-1.5 py-0.5 rounded opacity-70 font-mono tracking-tight"
          >
            {{ dateStr }}
          </span>
          <Code2 v-if="isCode" :size="12" class="text-accent opacity-50" />

          <span
            v-if="isModified"
            class="text-[9px] text-orange-400 font-bold bg-orange-400/10 px-1 rounded border border-orange-400/20"
          >
            EDITADO
          </span>
        </div>
        <div
          v-if="wasCopied"
          class="flex items-center gap-1 text-[10px] font-bold text-green-400 animate-pulse"
        >
          <Check :size="12" /> COPIADO
        </div>
      </div>

      <div
        v-if="showTools"
        class="mt-3 pt-2 border-t border-white/10 flex gap-2 overflow-x-auto pb-1"
        @click.stop
      >
        <button
          v-if="isModified"
          @click="resetContent"
          class="flex-shrink-0 flex items-center gap-1 px-2 py-1 rounded bg-orange-500/10 text-[10px] text-orange-400 hover:text-white hover:bg-orange-500 border border-orange-500/30 transition-colors"
          title="Restaurar Original"
        >
          <RotateCcw :size="10" /> Reset
        </button>

        <div
          v-if="isModified"
          class="w-[1px] bg-white/10 mx-1 flex-shrink-0"
        ></div>

        <button
          @click="applyTool('upper')"
          class="flex-shrink-0 flex items-center gap-1 px-2 py-1 rounded bg-[#11111b] text-[10px] text-gray-300 hover:text-white hover:bg-accent/20 border border-white/5 transition-colors"
        >
          <ArrowUpAZ :size="12" /> ABC
        </button>

        <button
          @click="applyTool('lower')"
          class="flex-shrink-0 flex items-center gap-1 px-2 py-1 rounded bg-[#11111b] text-[10px] text-gray-300 hover:text-white hover:bg-accent/20 border border-white/5 transition-colors"
        >
          <ArrowDownAZ :size="12" /> abc
        </button>

        <div class="w-[1px] bg-white/10 mx-1 flex-shrink-0"></div>

        <button
          @click="applyTool('json')"
          class="flex-shrink-0 p-1.5 rounded bg-[#11111b] text-gray-400 hover:text-accent border border-white/5"
          title="Format JSON"
        >
          <Braces :size="10" />
        </button>

        <button
          @click="applyTool('mini')"
          class="flex-shrink-0 p-1.5 rounded bg-[#11111b] text-gray-400 hover:text-accent border border-white/5"
          title="Minify"
        >
          <Minimize :size="10" />
        </button>
      </div>
    </div>
    <div
      class="absolute top-2 right-2 flex flex-col gap-1 transition-opacity duration-200"
      :class="
        selected || wasCopied || showTools
          ? 'opacity-100'
          : 'opacity-0 group-hover:opacity-100'
      "
    >
      <button
        @click.stop="emit('pin', clip.id)"
        class="p-1.5 rounded bg-bg text-subtext hover:text-warn hover:bg-overlay shadow-sm"
      >
        <Pin :size="14" :fill="clip.is_pinned ? 'currentColor' : 'none'" />
      </button>

      <button
        @click.stop="showTools = !showTools"
        class="p-1.5 rounded bg-bg text-subtext hover:text-accent hover:bg-overlay shadow-sm"
        :class="{ 'text-accent bg-accent/10': showTools }"
      >
        <Wand2 :size="14" />
      </button>

      <button
        @click.stop="emit('delete', clip.id)"
        class="p-1.5 rounded bg-bg text-subtext hover:text-red-400 hover:bg-overlay shadow-sm"
      >
        <Trash2 :size="14" />
      </button>
    </div>
  </div>
</template>
