<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import {
  GetClips,
  TogglePin,
  DeleteClip,
  CleanClipsOlderThan,
  CleanAllUnpinned,
  CleanTodayClips,
  WriteToClipboard,
  AddClip,
} from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import ClipCard from './components/ClipCard.vue';
import { Clock, Pin, List, AlignLeft, Code, X, Trash2, ChevronLeft, ChevronRight } from 'lucide-vue-next';
import type { Clip } from './types/Clip';
import type { app } from '../wailsjs/go/models';

// --- Estado ---
const activeTab = ref<'recent' | 'pinned'>('recent');
const clips = ref<Clip[]>([]);
const selectedIds = ref<Set<string>>(new Set());
const isCleanModalOpen = ref(false);
const isLoading = ref(true);

// --- Estado de Paginação ---
const currentPage = ref(1);
const pageSize = 50;
const totalPages = ref(1);
const totalItems = ref(0);

// --- Funções de Carregamento ---
const refreshClips = async () => {
  isLoading.value = true;
  clearSelection();
  try {
    // O backend espera `false` para 'recentes' e `true` para 'salvos'
    const isPinned = activeTab.value === 'pinned';
    const result: app.PaginatedClips | null = await GetClips(currentPage.value, pageSize, isPinned);

    if (result) {
      clips.value = result.clips || [];
      totalPages.value = result.total_pages;
      totalItems.value = result.total_items;
      currentPage.value = result.page;
    } else {
      clips.value = [];
      totalPages.value = 1;
      totalItems.value = 0;
    }
  } catch (e) {
    console.error("Erro ao carregar clips:", e);
    clips.value = [];
  } finally {
    isLoading.value = false;
  }
};

// --- Lifecycle e Watchers ---
onMounted(async () => {
  await refreshClips();

  EventsOn("clip-added", async () => {
    // Se estivermos na primeira página de recentes, atualiza para ver o novo item
    if (activeTab.value === 'recent' && currentPage.value === 1) {
      await refreshClips();
    }
  });

  EventsOn("clips-cleaned", async (count: number) => {
    console.log(`${count} clips antigos foram limpos.`);
    await refreshClips();
    // Poderia adicionar uma notificação para o usuário aqui
  });
});

watch(activeTab, async () => {
  currentPage.value = 1;
  await refreshClips();
});

// --- Ações de Clips ---
const toggleSelect = (id: string) => {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id);
  } else {
    selectedIds.value.add(id);
  }
};

const clearSelection = () => selectedIds.value.clear();

const mergeSelection = async (type: 'list' | 'paragraph' | 'code') => {
  // Filtra os clips selecionados e reverte para manter a ordem cronológica (mais antigo primeiro)
  const itemsToMerge = clips.value
    .filter(c => selectedIds.value.has(c.id))
    .map(c => c.content)
    .reverse();
  
  if (itemsToMerge.length === 0) return;

  let finalContent = "";
  switch (type) {
    case 'list': finalContent = itemsToMerge.map(t => `- ${t}`).join('\n'); break;
    case 'paragraph': finalContent = itemsToMerge.join('\n\n'); break;
    case 'code': finalContent = itemsToMerge.join('\n'); break;
  }

  await AddClip(finalContent);
  // O evento 'clip-added' vai disparar o refresh
};

const handlePin = async (id: string) => {
  await TogglePin(id);
  await refreshClips(); // Recarrega para refletir a mudança de aba
};

const handleDelete = async (id: string) => {
  await DeleteClip(id);
  // Se a página ficar vazia, volta para a anterior
  if (clips.value.length === 1 && currentPage.value > 1) {
    currentPage.value--;
  }
  await refreshClips();
};

const handleCopy = async (content: string) => {
  await WriteToClipboard(content);
};

// --- Ações de Limpeza e Paginação ---
const runClean = async (mode: 'today' | 'olderThan30Days' | 'allUnpinned') => {
  let count = 0;
  let actionDescription = '';

  try {
    switch (mode) {
      case 'today':
        if (!confirm("Deseja apagar todos os clips de hoje (exceto os salvos)?")) return;
        actionDescription = 'clips de hoje';
        count = await CleanTodayClips();
        break;
      case 'olderThan30Days':
        if (!confirm("Deseja apagar todos os clips com mais de 30 dias (exceto os salvos)?")) return;
        actionDescription = 'clips com mais de 30 dias';
        count = await CleanClipsOlderThan(30);
        break;
      case 'allUnpinned':
        if (!confirm("ATENÇÃO: Esta ação é irreversível. Deseja apagar TODOS os clips que não estão salvos?")) return;
        actionDescription = 'todos os clips não salvos';
        count = await CleanAllUnpinned();
        break;
    }
    alert(`${count} ${actionDescription} foram apagados.`);
  } catch (e) {
    alert(`Erro ao limpar clips: ${e}`);
  } finally {
    isCleanModalOpen.value = false;
  }
};

const goToPage = async (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
    await refreshClips();
  }
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
        <div class="flex items-center gap-4">
          <div class="text-[10px] text-gray-600 font-mono hidden sm:block opacity-60">
              {{ totalItems }} clips
          </div>
          <button @click="isCleanModalOpen = true" class="text-gray-600 hover:text-red-500 transition-colors" title="Opções de Limpeza">
            <Trash2 :size="16" />
          </button>
        </div>
    </div>

    <div class="flex-1 overflow-y-auto p-4 custom-scrollbar bg-surface/30 relative">
        <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-bg/50 z-20">
          <p>Carregando...</p>
        </div>
        <div v-else-if="clips.length > 0" class="flex flex-col gap-2 pb-20">
            <ClipCard 
              v-for="clip in clips" 
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
            <span class="text-sm font-medium">Nenhum item encontrado.</span>
        </div>
    </div>

    <!-- Barra de Ações de Seleção -->
    <Transition 
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-20 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-20 opacity-0"
    >
      <div v-if="selectedIds.size > 0" class="absolute bottom-20 left-1/2 -translate-x-1/2 z-50">
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

    <!-- Controles de Paginação -->
    <div v-if="totalPages > 1" class="flex items-center justify-center px-4 py-3 bg-bg border-t border-overlay z-10">
      <div class="flex items-center gap-2">
        <button 
          @click="goToPage(currentPage - 1)" 
          :disabled="currentPage === 1"
          class="p-2 rounded-md hover:bg-overlay disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <ChevronLeft :size="16" />
        </button>
        <span class="text-xs font-mono text-gray-400">
          Página {{ currentPage }} de {{ totalPages }}
        </span>
        <button 
          @click="goToPage(currentPage + 1)" 
          :disabled="currentPage === totalPages"
          class="p-2 rounded-md hover:bg-overlay disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <ChevronRight :size="16" />
        </button>
      </div>
    </div>

    <!-- Modal de Limpeza -->
    <div v-if="isCleanModalOpen" class="fixed inset-0 bg-black/60 z-40 flex items-center justify-center" @click.self="isCleanModalOpen = false">
        <div class="bg-surface rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
            <h3 class="text-lg font-bold mb-4">Opções de Limpeza</h3>
            <p class="text-sm text-subtle mb-6">Selecione uma opção para limpar os clips. Itens salvos (fixados) não serão afetados.</p>
            
            <div class="flex flex-col gap-3">
                <button @click="runClean('today')" class="w-full text-left p-3 bg-overlay hover:bg-overlay/70 rounded-lg transition-colors">Limpar clips de Hoje</button>
                <button @click="runClean('olderThan30Days')" class="w-full text-left p-3 bg-overlay hover:bg-overlay/70 rounded-lg transition-colors">Limpar clips com mais de 30 dias</button>
                <button @click="runClean('allUnpinned')" class="w-full text-left p-3 bg-red-500/20 hover:bg-red-500/40 text-red-400 rounded-lg transition-colors">
                  Limpar TUDO (exceto salvos)
                </button>
            </div>

            <div class="mt-6 text-right">
                <button @click="isCleanModalOpen = false" class="px-4 py-2 text-xs text-gray-400 hover:text-white rounded-md hover:bg-overlay">Cancelar</button>
            </div>
        </div>
    </div>

  </div>
</template>