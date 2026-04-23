<script lang="ts">
  import { appStore } from '../stores/app.svelte'
  import type { DiffLine } from '../types'

  let lastAnchor = $state<{ hunkIdx: number; lineIdx: number } | null>(null)

  function clickLine(hunkIdx: number, lineIdx: number, e: MouseEvent) {
    const line = appStore.diffFiles[0]?.hunks[hunkIdx]?.lines[lineIdx]
    if (!line || line.op === 'context') return
    // Shift-click → range
    if (e.shiftKey && lastAnchor && lastAnchor.hunkIdx === hunkIdx) {
      const [a, b] = [lastAnchor.lineIdx, lineIdx].sort((x, y) => x - y)
      for (let i = a; i <= b; i++) {
        const l = appStore.diffFiles[0].hunks[hunkIdx].lines[i]
        if (l && l.op !== 'context') {
          const k = `${hunkIdx}:${i}`
          const s = new Set(appStore.selectedLines)
          s.add(k)
          appStore.selectedLines = s
        }
      }
      return
    }
    appStore.toggleLine(hunkIdx, lineIdx)
    lastAnchor = { hunkIdx, lineIdx }
  }

  function lineClass(l: DiffLine): string {
    switch (l.op) {
      case 'add':
        return 'bg-[var(--color-diff-add-bg)] hover:brightness-110'
      case 'remove':
        return 'bg-[var(--color-diff-del-bg)] hover:brightness-110'
      default:
        return 'hover:bg-[var(--color-bg-soft)]'
    }
  }

  function stripeClass(l: DiffLine): string {
    switch (l.op) {
      case 'add':
        return 'bg-[var(--color-diff-add-stripe)]'
      case 'remove':
        return 'bg-[var(--color-diff-del-stripe)]'
      default:
        return 'bg-transparent'
    }
  }

  function selected(hi: number, li: number): boolean {
    return appStore.selectedLines.has(`${hi}:${li}`)
  }

  async function stageSelected() { await appStore.stageSelectedLines() }
  async function unstageSelected() { await appStore.unstageSelectedLines() }

  function selectAll(hi: number) { appStore.selectAllInHunk(hi) }

  function fmtNo(n: number): string {
    return n > 0 ? String(n) : ''
  }
</script>

{#if !appStore.selectedPath}
  <div class="flex-1 flex items-center justify-center text-fg-dim">ファイルを選択してね</div>
{:else}
  <div class="flex flex-col h-full overflow-hidden">
    <!-- Header -->
    <div class="flex items-center gap-2 px-3 py-2 border-b border-[var(--color-border)] bg-[var(--color-bg-soft)]">
      <span class="font-mono text-[13px] text-sky-300 truncate flex-1">
        {appStore.selectedPath}
        <span class="ml-2 text-[11px] text-fg-dim">({appStore.selectedStaged ? 'staged' : 'unstaged'})</span>
      </span>
      {#if appStore.hasLineSelection()}
        <span class="text-[11px] text-fg-muted">{appStore.selectedLines.size}行選択</span>
        {#if appStore.selectedStaged}
          <button
            type="button"
            class="px-2 py-0.5 text-[11px] rounded bg-amber-600 hover:bg-amber-500 text-white"
            onclick={unstageSelected}>選択をアンステージ</button>
        {:else}
          <button
            type="button"
            class="px-2 py-0.5 text-[11px] rounded bg-emerald-600 hover:bg-emerald-500 text-white"
            onclick={stageSelected}>選択をステージ</button>
        {/if}
        <button
          type="button"
          class="px-2 py-0.5 text-[11px] rounded bg-[var(--color-bg-softer)] hover:brightness-125 text-fg-muted"
          onclick={() => appStore.clearLineSelection()}>解除</button>
      {/if}
    </div>

    <!-- Body -->
    <div class="flex-1 overflow-auto font-mono leading-[1.5] no-select" style="user-select: text; font-size: var(--pik-font-size, 12px);">
      {#if appStore.loading}
        <div class="text-center text-fg-dim py-6">読込中…</div>
      {:else if appStore.diffFiles.length === 0}
        <div class="text-center text-fg-dim py-6">差分なし</div>
      {:else}
        {#each appStore.diffFiles as file}
          {#each file.hunks as hunk, hi}
            <!-- Hunk header -->
            <div class="flex bg-[var(--color-hunk-bg)] text-purple-300 px-2 py-0.5 group/hunk">
              <span class="flex-1 truncate">{hunk.header}</span>
              <button
                type="button"
                class="opacity-0 group-hover/hunk:opacity-100 text-[11px] hover:text-white"
                onclick={() => selectAll(hi)}>このハンクを全選択</button>
            </div>
            <!-- Lines -->
            {#each hunk.lines as line, li}
              {@const isSel = selected(hi, li)}
              <div
                class="flex {lineClass(line)} {isSel ? 'ring-1 ring-inset ring-white/50' : ''}"
                onclick={(e) => clickLine(hi, li, e)}
                role="row"
                onkeydown={(e) => e.key === ' ' && clickLine(hi, li, e as any)}
                tabindex={line.op === 'context' ? -1 : 0}
              >
                <!-- stripe -->
                <span class="w-[3px] shrink-0 {stripeClass(line)}"></span>
                <!-- gutter -->
                <span class="shrink-0 px-2 py-0 text-right text-fg-dim bg-[var(--color-gutter-bg)] select-none tabular-nums" style="min-width: 4em;">
                  {fmtNo(line.oldLineNo)}
                </span>
                <span class="shrink-0 px-2 py-0 text-right text-fg-dim bg-[var(--color-gutter-bg)] select-none tabular-nums" style="min-width: 4em;">
                  {fmtNo(line.newLineNo)}
                </span>
                <!-- sign -->
                <span class="shrink-0 w-4 text-center text-fg-dim select-none">
                  {line.op === 'add' ? '+' : line.op === 'remove' ? '-' : ' '}
                </span>
                <!-- content -->
                <span class="whitespace-pre flex-1 pr-2">{line.text}</span>
              </div>
            {/each}
          {/each}
          {#if file.binary}
            <div class="text-center text-fg-dim py-6">(バイナリファイル)</div>
          {/if}
        {/each}
      {/if}
    </div>
  </div>
{/if}
