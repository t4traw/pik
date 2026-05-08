<script lang="ts">
  import { appStore } from '../stores/app.svelte'
  import { t } from '../i18n/index.svelte'
  import type { ConflictFile, ConflictChoice } from '../types'
  import Icon from './Icon.svelte'

  type Choices = Record<number, ConflictChoice>

  let choicesByFile = $state<Record<string, Choices>>({})

  function close() {
    appStore.conflictsOpen = false
  }

  function onKeyDown(e: KeyboardEvent) {
    if (appStore.conflictsOpen && e.key === 'Escape') close()
  }

  function setChoice(path: string, regionIdx: number, choice: ConflictChoice) {
    if (!choicesByFile[path]) choicesByFile[path] = {}
    choicesByFile[path][regionIdx] = choice
  }

  function getChoice(path: string, regionIdx: number): ConflictChoice {
    return choicesByFile[path]?.[regionIdx] ?? null
  }

  function chosenCount(file: ConflictFile): number {
    const c = choicesByFile[file.path] ?? {}
    return file.regions.filter((_, i) => c[i] != null).length
  }

  function allChosen(file: ConflictFile): boolean {
    if (file.binary) return false
    return file.regions.length > 0 && chosenCount(file) === file.regions.length
  }

  // Walk the file's lines, replacing each conflict region with the user's pick.
  // startLine/endLine are 1-based and inclusive of the marker lines (<<<<<<<
  // and >>>>>>>), so we copy lines[i] for i in [0..startLine-2], emit the
  // chosen content, then resume from index endLine.
  function compose(file: ConflictFile): string {
    const choices = choicesByFile[file.path] ?? {}
    const out: string[] = []
    let i = 0
    for (let r = 0; r < file.regions.length; r++) {
      const region = file.regions[r]
      while (i < region.startLine - 1) {
        out.push(file.lines[i])
        i++
      }
      const choice = choices[r]
      if (choice === 'ours') out.push(...region.oursLines)
      else if (choice === 'theirs') out.push(...region.theirsLines)
      else if (choice === 'both') out.push(...region.oursLines, ...region.theirsLines)
      i = region.endLine
    }
    while (i < file.lines.length) {
      out.push(file.lines[i])
      i++
    }
    return out.join('\n') + '\n'
  }

  async function saveFile(file: ConflictFile) {
    if (!allChosen(file)) return
    await appStore.resolveFileWithContent(file.path, compose(file))
    delete choicesByFile[file.path]
    choicesByFile = { ...choicesByFile }
  }

  async function takeOurs(file: ConflictFile) {
    await appStore.resolveFileOurs(file.path)
  }

  async function takeTheirs(file: ConflictFile) {
    await appStore.resolveFileTheirs(file.path)
  }

  function isResolved(file: ConflictFile): boolean {
    // The store removes the file from conflictFiles after resolution, so any
    // file still in the list with zero regions is "binary, not yet resolved".
    return false
  }

  let allFilesResolved = $derived(appStore.conflictFiles.length === 0)
</script>

<svelte:window onkeydown={onKeyDown} />

{#if appStore.conflictsOpen}
  <div
    class="fixed inset-0 z-40 bg-black/70"
    role="button"
    tabindex="-1"
    aria-label={t('conflict.closeAria')}
    onclick={close}
    onkeydown={(e) => e.key === 'Enter' && close()}
  ></div>

  <div
    class="fixed inset-0 z-50 flex items-center justify-center pointer-events-none p-4"
    role="dialog"
    aria-modal="true"
    aria-label={t('conflict.title')}
  >
    <div class="pointer-events-auto w-full max-w-[860px] h-full flex flex-col rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-soft)] shadow-2xl">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)]">
        <div class="flex items-center gap-3 min-w-0">
          <span class="text-sm font-semibold">{t('conflict.title')}</span>
          {#if appStore.rebaseState.rebasing && appStore.rebaseState.total > 0}
            <span class="text-[11px] text-[var(--color-fg-muted)] tabular-nums">
              {t('conflict.rebaseProgress', { step: appStore.rebaseState.step, total: appStore.rebaseState.total })}
            </span>
          {:else if appStore.rebaseState.merging}
            <span class="text-[11px] text-[var(--color-fg-muted)]">{t('conflict.merging')}</span>
          {/if}
          {#if !allFilesResolved}
            <span class="text-[11px] text-amber-300">{t('conflict.subtitle', { count: appStore.conflictFiles.length })}</span>
          {:else}
            <span class="text-[11px] text-emerald-300">{t('conflict.allResolved')}</span>
          {/if}
        </div>
        <button
          type="button"
          aria-label={t('settings.close')}
          class="w-6 h-6 flex items-center justify-center rounded text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg-softer)]"
          onclick={close}
        >
          <Icon name="close" size={14} />
        </button>
      </div>

      <!-- Body -->
      <div class="flex-1 overflow-y-auto p-4 space-y-6">
        {#if allFilesResolved}
          <div class="text-center text-[var(--color-fg-muted)] py-12 text-[13px]">
            {t('conflict.allResolved')}
          </div>
        {/if}

        {#each appStore.conflictFiles as file (file.path)}
          <div class="rounded border border-[var(--color-border)] bg-[var(--color-bg)]">
            <div class="flex items-center justify-between px-3 py-2 border-b border-[var(--color-border)] bg-[var(--color-bg-soft)]">
              <span class="font-mono text-[13px] text-sky-300 truncate">{file.path}</span>
              <span class="shrink-0 text-[11px] text-[var(--color-fg-muted)] tabular-nums">
                {#if file.binary}
                  {t('conflict.binary')}
                {:else}
                  {chosenCount(file)} / {file.regions.length}
                {/if}
              </span>
            </div>

            {#if file.binary || file.regions.length === 0}
              <!-- Binary or unparseable: file-level resolution only. -->
              <div class="p-3 flex items-center gap-2">
                <button
                  type="button"
                  class="px-3 py-1 text-[12px] rounded bg-[var(--color-bg-softer)] hover:bg-[var(--color-accent)] hover:text-white"
                  onclick={() => takeOurs(file)}
                >
                  {t('conflict.useOurs')}
                </button>
                <button
                  type="button"
                  class="px-3 py-1 text-[12px] rounded bg-[var(--color-bg-softer)] hover:bg-[var(--color-accent)] hover:text-white"
                  onclick={() => takeTheirs(file)}
                >
                  {t('conflict.useTheirs')}
                </button>
              </div>
            {:else}
              <!-- Text file with parsed regions. -->
              <div class="p-3 space-y-4">
                {#each file.regions as region, ri (ri)}
                  {@const choice = getChoice(file.path, ri)}
                  <div class="rounded border border-[var(--color-border)]">
                    <div class="flex items-center justify-between px-2 py-1 bg-[var(--color-bg-softer)] border-b border-[var(--color-border)]">
                      <span class="text-[11px] text-[var(--color-fg-muted)]">
                        {t('conflict.region', { n: ri + 1, total: file.regions.length })}
                      </span>
                      <div class="flex items-center gap-1">
                        <button
                          type="button"
                          class="px-2 py-0.5 text-[11px] rounded {choice === 'ours' ? 'bg-emerald-600 text-white' : 'bg-[var(--color-bg)] text-[var(--color-fg-muted)] hover:text-white'}"
                          onclick={() => setChoice(file.path, ri, 'ours')}
                        >
                          {t('conflict.useOurs')}
                        </button>
                        <button
                          type="button"
                          class="px-2 py-0.5 text-[11px] rounded {choice === 'theirs' ? 'bg-emerald-600 text-white' : 'bg-[var(--color-bg)] text-[var(--color-fg-muted)] hover:text-white'}"
                          onclick={() => setChoice(file.path, ri, 'theirs')}
                        >
                          {t('conflict.useTheirs')}
                        </button>
                        <button
                          type="button"
                          class="px-2 py-0.5 text-[11px] rounded {choice === 'both' ? 'bg-emerald-600 text-white' : 'bg-[var(--color-bg)] text-[var(--color-fg-muted)] hover:text-white'}"
                          onclick={() => setChoice(file.path, ri, 'both')}
                        >
                          {t('conflict.useBoth')}
                        </button>
                      </div>
                    </div>

                    <!-- Side blocks. Highlight whichever the user picked. -->
                    <div class="font-mono text-[12px]" style="font-size: var(--pik-font-size, 12px);">
                      <div class="px-2 py-0.5 bg-[var(--color-bg-soft)] text-[10px] text-emerald-300 border-b border-[var(--color-border)]">
                        {t('conflict.ours', { label: region.oursLabel || 'HEAD' })}
                      </div>
                      <div class="{choice === 'ours' || choice === 'both' ? 'bg-[var(--color-diff-add-bg)]' : 'opacity-60'}">
                        {#each region.oursLines as line}
                          <div class="px-3 whitespace-pre">{line || ' '}</div>
                        {/each}
                        {#if region.oursLines.length === 0}
                          <div class="px-3 text-[var(--color-fg-dim)] italic">(empty)</div>
                        {/if}
                      </div>
                      <div class="px-2 py-0.5 bg-[var(--color-bg-soft)] text-[10px] text-amber-300 border-y border-[var(--color-border)]">
                        {t('conflict.theirs', { label: region.theirsLabel || '?' })}
                      </div>
                      <div class="{choice === 'theirs' || choice === 'both' ? 'bg-[var(--color-diff-add-bg)]' : 'opacity-60'}">
                        {#each region.theirsLines as line}
                          <div class="px-3 whitespace-pre">{line || ' '}</div>
                        {/each}
                        {#if region.theirsLines.length === 0}
                          <div class="px-3 text-[var(--color-fg-dim)] italic">(empty)</div>
                        {/if}
                      </div>
                    </div>
                  </div>
                {/each}

                <div class="flex items-center justify-end gap-2 pt-1">
                  <span class="text-[11px] text-[var(--color-fg-muted)]">
                    {#if allChosen(file)}
                      {t('conflict.allChosen')}
                    {:else}
                      {t('conflict.unresolvedRegions', { n: file.regions.length - chosenCount(file) })}
                    {/if}
                  </span>
                  <button
                    type="button"
                    class="px-3 py-1 text-[12px] rounded bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white font-semibold disabled:opacity-40"
                    disabled={!allChosen(file)}
                    onclick={() => saveFile(file)}
                  >
                    {t('conflict.save')}
                  </button>
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>

      <!-- Footer: rebase / merge controls. -->
      {#if appStore.rebaseState.rebasing || appStore.rebaseState.merging}
        <div class="flex items-center justify-between gap-2 px-4 py-2 border-t border-[var(--color-border)] bg-[var(--color-bg-soft)]">
          <span class="text-[11px] text-[var(--color-fg-muted)]">
            {#if !allFilesResolved}
              {t('conflict.someUnresolved', { n: appStore.conflictFiles.length })}
            {:else}
              {t('conflict.allResolved')}
            {/if}
          </span>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="px-3 py-1 text-[12px] rounded bg-[var(--color-bg-softer)] hover:bg-rose-700 hover:text-white text-[var(--color-fg)]"
              onclick={() => appStore.abortRebase()}
            >
              {t('conflict.abortRebase')}
            </button>
            <button
              type="button"
              class="px-3 py-1 text-[12px] rounded bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white font-semibold disabled:opacity-40"
              disabled={!allFilesResolved}
              onclick={() => appStore.continueRebase()}
            >
              {t('conflict.continueRebase')}
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}
