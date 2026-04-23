<script lang="ts">
  import type { FileStatus } from '../types'
  import { appStore } from '../stores/app.svelte'

  let dragOverTarget = $state<'staged' | 'changes' | null>(null)

  function badgeFor(f: FileStatus, staged: boolean): { char: string; color: string } {
    const code = staged ? f.IndexStatus : f.Untracked ? 63 : f.WorkStatus // '?' = 63
    switch (String.fromCharCode(code)) {
      case 'M': return { char: 'M', color: 'text-amber-400' }
      case 'A': return { char: 'A', color: 'text-emerald-400' }
      case 'D': return { char: 'D', color: 'text-rose-400' }
      case 'R': return { char: 'R', color: 'text-sky-300' }
      case 'C': return { char: 'C', color: 'text-sky-300' }
      case '?': return { char: 'U', color: 'text-emerald-400' }
      default: return { char: String.fromCharCode(code) || '·', color: 'text-fg-dim' }
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

<div class="flex flex-col h-full bg-[var(--color-bg-soft)] overflow-hidden text-[13px]">
  <!-- STAGED -->
  <div
    class="flex items-center justify-between px-3 py-1.5 bg-[var(--color-bg-softer)] border-b border-[var(--color-border)] {dragOverTarget === 'staged' ? 'ring-2 ring-inset ring-[var(--color-accent)]' : ''}"
    ondragover={(e) => onDragOver(e, 'staged')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'staged')}
    role="group"
  >
    <span class="text-[11px] font-semibold tracking-wide text-fg-muted">
      STAGED CHANGES ({appStore.stagedFiles.length})
    </span>
    {#if appStore.stagedFiles.length > 0}
      <button
        type="button"
        class="text-fg-muted hover:text-fg cursor-pointer"
        title="すべてアンステージ"
        onclick={() => appStore.unstageAll()}>↶ all</button>
    {/if}
  </div>
  <div
    class="overflow-y-auto {dragOverTarget === 'staged' ? 'bg-[var(--color-selected)]/10' : ''}"
    ondragover={(e) => onDragOver(e, 'staged')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'staged')}
    role="list"
  >
    {#each appStore.stagedFiles as f (f.Path)}
      {@const b = badgeFor(f, true)}
      {@const selected = appStore.selectedPath === f.Path && appStore.selectedStaged}
      <div
        class="group flex items-center px-2 py-1 cursor-pointer hover:bg-[var(--color-bg-softer)] {selected ? 'bg-[var(--color-selected)]' : ''}"
        draggable="true"
        ondragstart={(e) => onDragStart(e, f.Path, true)}
        role="listitem"
        onclick={() => appStore.selectFile(f, true)}
        onkeydown={(e) => e.key === 'Enter' && appStore.selectFile(f, true)}
        tabindex="0"
      >
        <span class="w-4 font-mono font-bold text-center {b.color}">{b.char}</span>
        <span class="ml-2 truncate flex-1">{basename(f.Path)}</span>
        <span class="ml-2 text-[11px] text-fg-dim truncate">{dirname(f.Path)}</span>
        <button
          type="button"
          class="ml-2 opacity-0 group-hover:opacity-100 hover:text-fg text-fg-muted"
          title="アンステージ"
          onclick={(e) => { e.stopPropagation(); appStore.unstage(f.Path) }}>↶</button>
      </div>
    {/each}
  </div>

  <!-- CHANGES -->
  <div
    class="flex items-center justify-between px-3 py-1.5 bg-[var(--color-bg-softer)] border-b border-t border-[var(--color-border)] {dragOverTarget === 'changes' ? 'ring-2 ring-inset ring-[var(--color-accent)]' : ''}"
    ondragover={(e) => onDragOver(e, 'changes')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'changes')}
    role="group"
  >
    <span class="text-[11px] font-semibold tracking-wide text-fg-muted">
      CHANGES ({appStore.unstagedFiles.length})
    </span>
    {#if appStore.unstagedFiles.length > 0}
      <button
        type="button"
        class="text-fg-muted hover:text-fg cursor-pointer"
        title="すべてステージ"
        onclick={() => appStore.stageAll()}>+ all</button>
    {/if}
  </div>
  <div
    class="flex-1 overflow-y-auto {dragOverTarget === 'changes' ? 'bg-[var(--color-selected)]/10' : ''}"
    ondragover={(e) => onDragOver(e, 'changes')}
    ondragleave={onDragLeave}
    ondrop={(e) => onDrop(e, 'changes')}
    role="list"
  >
    {#each appStore.unstagedFiles as f (f.Path)}
      {@const b = badgeFor(f, false)}
      {@const selected = appStore.selectedPath === f.Path && !appStore.selectedStaged}
      <div
        class="group flex items-center px-2 py-1 cursor-pointer hover:bg-[var(--color-bg-softer)] {selected ? 'bg-[var(--color-selected)]' : ''}"
        draggable="true"
        ondragstart={(e) => onDragStart(e, f.Path, false)}
        role="listitem"
        onclick={() => appStore.selectFile(f, false)}
        onkeydown={(e) => e.key === 'Enter' && appStore.selectFile(f, false)}
        tabindex="0"
      >
        <span class="w-4 font-mono font-bold text-center {b.color}">{b.char}</span>
        <span class="ml-2 truncate flex-1">{basename(f.Path)}</span>
        <span class="ml-2 text-[11px] text-fg-dim truncate">{dirname(f.Path)}</span>
        <button
          type="button"
          class="ml-1 opacity-0 group-hover:opacity-100 hover:text-rose-400 text-fg-muted"
          title={f.Untracked ? '削除' : '変更破棄'}
          onclick={(e) => { e.stopPropagation(); discardFile(f) }}>{f.Untracked ? '🗑' : '↶'}</button>
        <button
          type="button"
          class="ml-1 opacity-0 group-hover:opacity-100 hover:text-emerald-400 text-fg-muted"
          title="ステージ"
          onclick={(e) => { e.stopPropagation(); appStore.stage(f.Path) }}>+</button>
      </div>
    {/each}

    {#if appStore.files.length === 0}
      <div class="text-center text-fg-dim py-6 text-[12px]">変更なし</div>
    {/if}
  </div>
</div>
