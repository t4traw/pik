<script lang="ts">
  import type { FileStatus } from '../types'
  import { appStore } from '../stores/app.svelte'
  import Icon from './Icon.svelte'

  let dragOverTarget = $state<'staged' | 'changes' | null>(null)
  let listRoot = $state<HTMLElement | null>(null)

  $effect(() => {
    const path = appStore.selectedPath
    const staged = appStore.selectedStaged
    if (!path || !listRoot) return
    const sel = `[data-path="${CSS.escape(path)}"][data-staged="${staged ? '1' : '0'}"]`
    const el = listRoot.querySelector(sel) as HTMLElement | null
    el?.scrollIntoView({ block: 'nearest' })
  })

  function badgeFor(f: FileStatus, staged: boolean): { char: string; color: string } {
    const code = staged ? f.IndexStatus : f.Untracked ? 63 : f.WorkStatus // '?' = 63
    switch (String.fromCharCode(code)) {
      case 'M': return { char: 'M', color: 'text-amber-400' }
      case 'A': return { char: 'A', color: 'text-emerald-400' }
      case 'D': return { char: 'D', color: 'text-rose-400' }
      case 'R': return { char: 'R', color: 'text-sky-300' }
      case 'C': return { char: 'C', color: 'text-sky-300' }
      case '?': return { char: 'U', color: 'text-emerald-400' }
      default: return { char: String.fromCharCode(code) || '·', color: 'text-[var(--color-fg-dim)]' }
    }
  }

  function dirname(p: string): string {
    const i = p.lastIndexOf('/')
    return i >= 0 ? p.slice(0, i) : ''
  }

  function basename(p: string): string {
    const i = p.lastIndexOf('/')
    return i >= 0 ? p.slice(i + 1) : p
  }

  function onDragStart(e: DragEvent, path: string, staged: boolean) {
    if (!e.dataTransfer) return
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('application/x-pik-file', JSON.stringify({ path, staged }))
  }

  function onDragOver(e: DragEvent, target: 'staged' | 'changes') {
    if (!e.dataTransfer) return
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    dragOverTarget = target
  }

  function onDragLeave() {
    dragOverTarget = null
  }

  async function onDrop(e: DragEvent, target: 'staged' | 'changes') {
    e.preventDefault()
    dragOverTarget = null
    const raw = e.dataTransfer?.getData('application/x-pik-file')
    if (!raw) return
    const { path, staged } = JSON.parse(raw) as { path: string; staged: boolean }
    if (target === 'staged' && !staged) await appStore.stage(path)
    if (target === 'changes' && staged) await appStore.unstage(path)
  }

  async function discardFile(f: FileStatus) {
    const msg = f.Untracked
      ? `未追跡ファイルを削除する？\n${f.Path}`
      : `変更を破棄する？\n${f.Path}`
    if (confirm(msg)) await appStore.discard(f.Path, f.Untracked)
  }
</script>

<div bind:this={listRoot} class="flex flex-col h-full bg-[var(--color-bg-soft)] overflow-hidden text-[13px]">
  <!-- STAGED header + drop zone -->
  <div
    role="group"
    ondragover={(e) => onDragOver(e, 'staged')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'staged')}
    class="flex items-center justify-between px-3 h-8 bg-[var(--color-bg-softer)] border-b border-[var(--color-border)] {dragOverTarget === 'staged' ? 'ring-2 ring-inset ring-[var(--color-accent)]' : ''}"
  >
    <span class="text-[11px] font-semibold tracking-wider text-[var(--color-fg-muted)]">
      STAGED CHANGES ({appStore.stagedFiles.length})
    </span>
    {#if appStore.stagedFiles.length > 0}
      <button
        type="button"
        aria-label="すべてアンステージ"
        class="flex items-center gap-1 h-6 px-2 rounded text-[11px] text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg)] transition-colors"
        onclick={() => appStore.unstageAll()}>
        <Icon name="undo" size={13} />
        <span>all</span>
      </button>
    {/if}
  </div>

  <div
    role="list"
    ondragover={(e) => onDragOver(e, 'staged')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'staged')}
    class="overflow-y-auto {dragOverTarget === 'staged' ? 'bg-[var(--color-selected)]/10' : ''}"
  >
    {#each appStore.stagedFiles as f (f.Path)}
      {@const b = badgeFor(f, true)}
      {@const selected = appStore.selectedPath === f.Path && appStore.selectedStaged}
      <div
        class="group flex items-center pl-2 pr-1 h-7 cursor-pointer hover:bg-[var(--color-bg-softer)] {selected ? 'bg-[var(--color-selected)]' : ''}"
        draggable="true"
        data-path={f.Path}
        data-staged="1"
        ondragstart={(e) => onDragStart(e, f.Path, true)}
        role="listitem"
        onclick={() => appStore.selectFile(f, true)}
        onkeydown={(e) => e.key === 'Enter' && appStore.selectFile(f, true)}
        tabindex="0"
      >
        <span class="w-4 font-mono font-bold text-center text-[12px] {b.color}">{b.char}</span>
        <span class="ml-2 truncate flex-1 text-[var(--color-fg)]">{basename(f.Path)}</span>
        <span class="ml-2 text-[11px] text-[var(--color-fg-dim)] truncate max-w-[40%]">{dirname(f.Path)}</span>
        <button
          type="button"
          aria-label="アンステージ"
          class="ml-1 w-6 h-6 flex items-center justify-center rounded opacity-0 group-hover:opacity-100 text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg)]"
          onclick={(e) => { e.stopPropagation(); appStore.unstage(f.Path) }}>
          <Icon name="undo" size={14} />
        </button>
      </div>
    {/each}
  </div>

  <!-- CHANGES header + drop zone -->
  <div
    role="group"
    ondragover={(e) => onDragOver(e, 'changes')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'changes')}
    class="flex items-center justify-between px-3 h-8 bg-[var(--color-bg-softer)] border-b border-t border-[var(--color-border)] {dragOverTarget === 'changes' ? 'ring-2 ring-inset ring-[var(--color-accent)]' : ''}"
  >
    <span class="text-[11px] font-semibold tracking-wider text-[var(--color-fg-muted)]">
      CHANGES ({appStore.unstagedFiles.length})
    </span>
    {#if appStore.unstagedFiles.length > 0}
      <button
        type="button"
        aria-label="すべてステージ"
        class="flex items-center gap-1 h-6 px-2 rounded text-[11px] text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg)] transition-colors"
        onclick={() => appStore.stageAll()}>
        <Icon name="plus" size={13} />
        <span>all</span>
      </button>
    {/if}
  </div>
  <div
    role="list"
    ondragover={(e) => onDragOver(e, 'changes')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'changes')}
    class="flex-1 overflow-y-auto {dragOverTarget === 'changes' ? 'bg-[var(--color-selected)]/10' : ''}"
  >
    {#each appStore.unstagedFiles as f (f.Path)}
      {@const b = badgeFor(f, false)}
      {@const selected = appStore.selectedPath === f.Path && !appStore.selectedStaged}
      <div
        class="group flex items-center pl-2 pr-1 h-7 cursor-pointer hover:bg-[var(--color-bg-softer)] {selected ? 'bg-[var(--color-selected)]' : ''}"
        draggable="true"
        data-path={f.Path}
        data-staged="0"
        ondragstart={(e) => onDragStart(e, f.Path, false)}
        role="listitem"
        onclick={() => appStore.selectFile(f, false)}
        onkeydown={(e) => e.key === 'Enter' && appStore.selectFile(f, false)}
        tabindex="0"
      >
        <span class="w-4 font-mono font-bold text-center text-[12px] {b.color}">{b.char}</span>
        <span class="ml-2 truncate flex-1 text-[var(--color-fg)]">{basename(f.Path)}</span>
        <span class="ml-2 text-[11px] text-[var(--color-fg-dim)] truncate max-w-[40%]">{dirname(f.Path)}</span>
        <button
          type="button"
          aria-label={f.Untracked ? '削除' : '変更破棄'}
          class="ml-1 w-6 h-6 flex items-center justify-center rounded opacity-0 group-hover:opacity-100 text-[var(--color-fg-muted)] hover:text-rose-300 hover:bg-[var(--color-bg)]"
          onclick={(e) => { e.stopPropagation(); discardFile(f) }}>
          <Icon name={f.Untracked ? 'trash' : 'undo'} size={14} />
        </button>
        <button
          type="button"
          aria-label="ステージ"
          class="ml-0.5 w-6 h-6 flex items-center justify-center rounded opacity-0 group-hover:opacity-100 text-[var(--color-fg-muted)] hover:text-emerald-300 hover:bg-[var(--color-bg)]"
          onclick={(e) => { e.stopPropagation(); appStore.stage(f.Path) }}>
          <Icon name="plus" size={14} />
        </button>
      </div>
    {/each}

    {#if appStore.files.length === 0}
      <div class="text-center text-[var(--color-fg-dim)] py-6 text-[12px]">変更なし</div>
    {/if}
  </div>
</div>
