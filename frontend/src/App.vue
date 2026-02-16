<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { GetClips, TogglePin, DeleteClip, WriteToClipboard, AddClip } from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import ClipCard from './components/ClipCard.vue';
import { Clock, Pin, List, AlignLeft, Code, X } from 'lucide-vue-next';
import type { Clip } from './types/Clip';

// --- Estado ---
const activeTab = ref<'recent' | 'pinned'>('recent');
const clips = ref<Clip[]>([]);
const selectedIds = ref<Set<string>>(new Set());

// --- Lifecycle ---
onMounted(async () => {
  await refreshClips();
  EventsOn("clip-added", (newClip: Clip) => {
    clips.value.unshift(newClip);
  });
});

const refreshClips = async () => {
  try {
    const result = await GetClips();
    clips.value = result || [];
  } catch (e) {
    console.error("Erro ao carregar clips:", e);
  }
};

// --- Computados ---
const filteredClips = computed(() => {
  if (activeTab.value === 'pinned') return clips.value.filter(c => c.is_pinned);
  return clips.value.filter(c => !c.is_pinned);
});

// --- Ações ---
const toggleSelect = (id: string) => {
  if (selectedIds.value.has(id)) selectedIds.value.delete(id);
  else selectedIds.value.add(id);
};

const clearSelection = () => selectedIds.value.clear();

const mergeSelection = async (type: 'list' | 'paragraph' | 'code') => {
  const itemsToMerge = clips.value.filter(c => selectedIds.value.has(c.id)).map(c => c.content);
  itemsToMerge.reverse();
  if (itemsToMerge.length === 0) return;

  let finalContent = "";
  switch (type) {
    case 'list': finalContent = itemsToMerge.map(t => `- ${t}`).join('\n'); break;
    case 'paragraph': finalContent = itemsToMerge.join('\n\n'); break;
    case 'code': finalContent = itemsToMerge.join('\n'); break;
  }

  await AddClip(finalContent);
  clearSelection();
};

const handlePin = async (id: string) => {
  await TogglePin(id);
  const item = clips.value.find(c => c.id === id);
  if (item) item.is_pinned = !item.is_pinned;
};

const handleDelete = async (id: string) => {
  await DeleteClip(id);
  clips.value = clips.value.filter(c => c.id !== id);
  selectedIds.value.delete(id);
};

const handleCopy = async (content: string) => {
  await WriteToClipboard(content);
};
</script>

<template>
  <div class="h-screen w-full bg-bg text-text flex flex-col font-sans overflow-hidden">
    
    <div class="flex items-center justify-between px-4 py-3 bg-bg border-b border-overlay shadow-sm z-10">
        <div class="flex space-x-2 bg-[#11111b] p-1 rounded-lg">
            <button @click="activeTab = 'recent'" class="flex items-center gap-2 text-xs font-bold px-3 py-1.5 rounded-md transition-colors" :class="activeTab === 'recent' ? 'bg-overlay text-white shadow' : 'text-gray-500 hover:text-gray-300'">
                <Clock :size="14" /> Recentes
            </button>
            <button @click="activeTab = 'pinned'" class="flex items-center gap-2 text-xs font-bold px-3 py-1.5 rounded-md transition-colors" :class="activeTab === 'pinned' ? 'bg-overlay text-yellow-500 shadow' : 'text-gray-500 hover:text-gray-300'">
                <Pin :size="14" /> Salvos
            </button>
        </div>
        <div class="text-[10px] text-gray-600 font-mono hidden sm:block opacity-60">
            {{ clips.length }} clips
        </div>
    </div>

    <div class="flex-1 overflow-y-auto p-4 custom-scrollbar bg-surface/30">
        <div v-if="filteredClips.length > 0" class="flex flex-col gap-2 pb-20">
            <ClipCard 
              v-for="clip in filteredClips" 
              :key="clip.id" 
              :clip="clip"
              :selected="selectedIds.has(clip.id)"
              @toggle-select="toggleSelect"
              @pin="handlePin"
              @delete="handleDelete"
              @copy="handleCopy"
            />
        </div>
        
        <div v-else class="flex flex-col items-center justify-center h-full text-gray-500 opacity-50">
            <span class="text-sm font-medium">Lista Vazia</span>
        </div>
    </div>

    <Transition 
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-20 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-20 opacity-0"
    >
      <div v-if="selectedIds.size > 0" class="absolute bottom-6 left-1/2 -translate-x-1/2 z-50">
        <div class="flex items-center gap-1 p-1.5 bg-[#11111b] border border-accent/30 rounded-full shadow-2xl shadow-black ring-1 ring-black/50">
          <div class="px-3 text-xs font-bold text-accent border-r border-white/10">{{ selectedIds.size }}</div>
          <button @click="mergeSelection('list')" class="p-2 rounded-full hover:bg-overlay text-text" title="Lista"><List :size="16" /></button>
          <button @click="mergeSelection('paragraph')" class="p-2 rounded-full hover:bg-overlay text-text" title="Texto"><AlignLeft :size="16" /></button>
          <button @click="mergeSelection('code')" class="p-2 rounded-full hover:bg-overlay text-text" title="Código"><Code :size="16" /></button>
          <div class="w-[1px] h-4 bg-white/10 mx-1"></div>
          <button @click="clearSelection" class="p-2 rounded-full hover:bg-red-500/20 text-gray-400 hover:text-red-400"><X :size="16" /></button>
        </div>
      </div>
    </Transition>
  </div>
</template>