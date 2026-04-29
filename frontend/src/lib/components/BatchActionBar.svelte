<script lang="ts">
  import { _ } from 'svelte-i18n';
  import { Button } from "$lib/components/ui/button/index.js";
  import { fly } from 'svelte/transition';

  let { 
    selectedCount = 0, 
    onClear = () => {} 
  } = $props<{
    selectedCount: number;
    onClear: () => void;
  }>();
</script>

{#if selectedCount > 0}
  <div 
    transition:fly={{ y: -50, duration: 400 }}
    class="fixed top-6 md:top-10 left-1/2 -translate-x-1/2 z-[100] flex items-center gap-3 md:gap-6 px-4 md:px-8 py-3 md:py-4 bg-white border border-slate-200 rounded-2xl md:rounded-3xl shadow-[0_30px_60px_rgba(0,0,0,0.25)] ring-1 ring-slate-950/5 w-[90vw] md:w-auto min-w-fit max-w-[95vw]"
  >
    <div class="flex items-center gap-2 md:gap-4 pr-3 md:pr-6 border-r border-slate-100">
      <div class="flex h-8 w-8 md:h-10 md:w-10 items-center justify-center rounded-xl md:rounded-2xl bg-primary text-white font-black text-xs md:text-sm shadow-lg shadow-primary/30 rotate-3">
        {selectedCount}
      </div>
      <div class="flex flex-col">
        <span class="text-[8px] md:text-[10px] font-black uppercase tracking-widest text-slate-400 leading-tight">
          {$_('common.selection')}
        </span>
        <span class="text-[10px] md:text-xs font-bold text-slate-900 hidden sm:inline">
          {$_('common.items_selected')}
        </span>
      </div>
    </div>

    <div class="flex items-center gap-2 md:gap-3 flex-1 justify-end sm:justify-start">
      <slot></slot>
      
      <div class="h-6 w-[1px] bg-slate-100 mx-1 md:mx-2 hidden sm:block"></div>

      <Button 
        variant="ghost" 
        size="sm" 
        onclick={onClear}
        class="h-8 md:h-10 px-2 md:px-4 text-[10px] md:text-xs font-black uppercase tracking-widest text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-lg md:rounded-xl transition-all"
      >
        {$_('common.cancel')}
      </Button>
    </div>
  </div>
{/if}
