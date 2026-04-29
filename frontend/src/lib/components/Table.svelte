<script lang="ts">
  import { _ } from 'svelte-i18n';
  import * as Table from "$lib/components/ui/table/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Search } from "lucide-svelte";

  let { 
    data = [], 
    columns = [], 
    actions = [], 
    loading = false,
    enableSelection = false,
    selectedIds = $bindable(new Set<number>())
  } = $props<{
    data: any[];
    columns: { key: string; label: string; render?: (val: any, row: any) => string }[];
    actions?: { label: string; class?: string; onClick: (item: any) => void }[];
    loading?: boolean;
    enableSelection?: boolean;
    selectedIds?: Set<number>;
  }>();
  
  let search = $state('');
  
  let filteredData = $derived(data.filter(item => {
    return Object.values(item).some(val => 
      String(val).toLowerCase().includes(search.toLowerCase())
    );
  }));

  function toggleSelectAll() {
    if (selectedIds.size === filteredData.length && filteredData.length > 0) {
      selectedIds = new Set();
    } else {
      selectedIds = new Set(filteredData.map(d => d.id));
    }
  }

  function toggleItem(id: number) {
    if (selectedIds.has(id)) {
      selectedIds.delete(id);
    } else {
      selectedIds.add(id);
    }
    selectedIds = new Set(selectedIds);
  }
</script>

<div class="flex flex-col gap-4">
  <div class="flex justify-between items-center gap-4">
    <div class="relative w-full max-w-sm">
      <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        placeholder={$_('common.search_placeholder')}
        bind:value={search}
        class="pl-10"
      />
    </div>
    <slot name="header-right"></slot>
  </div>

  <div class="rounded-md border bg-card shadow-lg overflow-hidden transition-all duration-300">
    <Table.Root>
      <Table.Header class="bg-muted/50">
        <Table.Row>
          {#if enableSelection}
            <Table.Head class="w-12 px-6">
              <input 
                type="checkbox" 
                class="h-4 w-4 rounded border-slate-300 text-primary focus:ring-primary cursor-pointer transition-all"
                checked={selectedIds.size === filteredData.length && filteredData.length > 0}
                onchange={toggleSelectAll}
              />
            </Table.Head>
          {/if}
          {#each columns as col}
            <Table.Head class="h-12 px-6 text-xs font-bold uppercase tracking-wider">{col.label}</Table.Head>
          {/each}
          {#if actions.length > 0}
            <Table.Head class="h-12 px-6 text-right text-xs font-bold uppercase tracking-wider">{$_('common.actions')}</Table.Head>
          {/if}
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#if loading}
          {#each Array(3) as _}
            <Table.Row class="animate-pulse">
              {#if enableSelection}
                <Table.Cell class="p-6"><div class="h-4 w-4 rounded bg-muted"></div></Table.Cell>
              {/if}
              {#each Array(columns.length + (actions.length ? 1 : 0)) as _}
                <Table.Cell class="p-6 text-center">
                  <div class="h-4 w-full rounded bg-muted"></div>
                </Table.Cell>
              {/each}
            </Table.Row>
          {/each}
        {:else}
          {#each filteredData as item}
            <Table.Row class="transition-colors hover:bg-muted/30 {selectedIds.has(item.id) ? 'bg-primary/5' : ''}">
              {#if enableSelection}
                <Table.Cell class="p-6">
                  <input 
                    type="checkbox" 
                    class="h-4 w-4 rounded border-slate-300 text-primary focus:ring-primary cursor-pointer transition-all"
                    checked={selectedIds.has(item.id)}
                    onchange={() => toggleItem(item.id)}
                  />
                </Table.Cell>
              {/if}
              {#each columns as col}
                <Table.Cell class="p-6">
                  {#if col.render}
                    {@html col.render(item[col.key], item)}
                  {:else}
                    {item[col.key]}
                  {/if}
                </Table.Cell>
              {/each}
              {#if actions.length > 0}
                <Table.Cell class="p-6 text-right">
                  <div class="flex justify-end gap-2">
                    {#each actions as action}
                      <Button 
                        variant="secondary"
                        size="sm"
                        class="h-8 px-3 text-xs font-bold {action.class}"
                        onclick={() => action.onClick(item)}
                      >
                        {action.label}
                      </Button>
                    {/each}
                  </div>
                </Table.Cell>
              {/if}
            </Table.Row>
          {:else}
            <Table.Row>
              <Table.Cell colspan={columns.length + (actions.length ? 1 : 0) + (enableSelection ? 1 : 0)} class="h-32 text-center text-muted-foreground italic">
                {$_('common.no_results')}
              </Table.Cell>
            </Table.Row>
          {/each}
        {/if}
      </Table.Body>
    </Table.Root>
  </div>
</div>
