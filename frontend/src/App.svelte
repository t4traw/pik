<script lang="ts">
  import { onMount } from 'svelte'
  import { appStore } from './lib/stores/app.svelte'
  import FileList from './lib/components/FileList.svelte'
  import DiffView from './lib/components/DiffView.svelte'
  import CommitBox from './lib/components/CommitBox.svelte'

  onMount(() => {
    appStore.refresh()
  })
</script>

<div class="flex flex-col h-full">
  <!-- Title bar -->
  <div class="flex items-center gap-2 px-3 py-1.5 bg-[var(--color-bg-soft)] border-b border-[var(--color-border)] text-[12px] select-none" style="--wails-draggable: drag;">
    <span class="text-sky-300 font-semibold">● {appStore.info.branch}</span>
    <span class="text-fg-dim truncate flex-1">{appStore.info.root}</span>
    <button
      type="button"
      class="text-fg-muted hover:text-fg px-2 cursor-pointer"
      title="Refresh"
      onclick={() => appStore.refresh()}>⟳</button>
  </div>

  <!-- Main split -->
  <div class="flex-1 grid overflow-hidden" style="grid-template-columns: minmax(240px, 320px) 1fr;">
    <!-- Left column: file list + commit box -->
    <div class="flex flex-col border-r border-[var(--color-border)] overflow-hidden">
      <div class="flex-1 overflow-hidden">
        <FileList />
      </div>
      <CommitBox />
    </div>

    <!-- Right column: diff -->
    <div class="flex flex-col overflow-hidden bg-[var(--color-bg)]">
      <DiffView />
    </div>
  </div>
</div>
