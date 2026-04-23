<script lang="ts">
  import { onMount } from 'svelte'
  import { appStore } from './lib/stores/app.svelte'
  import FileList from './lib/components/FileList.svelte'
  import DiffView from './lib/components/DiffView.svelte'
  import CommitBox from './lib/components/CommitBox.svelte'
  import Icon from './lib/components/Icon.svelte'

  onMount(() => {
    appStore.refresh()
    const onFocus = () => appStore.refresh()
    window.addEventListener('focus', onFocus)
    return () => window.removeEventListener('focus', onFocus)
  })
</script>

<div class="flex flex-col h-full">
  <!-- Title bar (drag region, matches macOS traffic-light layout) -->
  <div
    class="flex items-center gap-3 pr-2 bg-[var(--color-bg-soft)] border-b border-[var(--color-border)] text-[12px] select-none"
    style="--wails-draggable: drag; padding-left: 80px; height: 28px;"
  >
    <span class="flex items-center gap-1.5 text-sky-300 font-semibold shrink-0">
      <Icon name="branch" size={14} />
      <span>{appStore.info.branch || '—'}</span>
    </span>
    <span class="text-[var(--color-fg-muted)] truncate flex-1 min-w-0">{appStore.info.root}</span>
    <button
      type="button"
      aria-label="Refresh"
      class="shrink-0 w-7 h-7 flex items-center justify-center rounded text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg-softer)] transition-colors"
      style="--wails-draggable: no-drag;"
      onclick={() => appStore.refresh()}>
      <Icon name="refresh" size={15} />
    </button>
  </div>

  <!-- Main split -->
  <div class="flex-1 grid overflow-hidden" style="grid-template-columns: minmax(260px, 340px) 1fr;">
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
