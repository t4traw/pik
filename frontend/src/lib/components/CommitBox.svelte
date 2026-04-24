<script lang="ts">
  import { appStore } from '../stores/app.svelte'
  import Icon from './Icon.svelte'

  function onKeyDown(e: KeyboardEvent) {
    // Cmd/Ctrl + Enter submits, without competing with IME confirm Enter.
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      appStore.commit()
    }
  }
</script>

<div class="border-t border-[var(--color-border)] bg-[var(--color-bg-soft)] p-2 flex flex-col gap-2">
  <textarea
    id="pik-panel-commit"
    class="w-full resize-none bg-[var(--color-bg)] text-fg text-[13px] p-2 rounded border border-[var(--color-border)] focus:outline-none focus:border-[var(--color-accent)] font-sans"
    placeholder="コミットメッセージ ( ⌘↵ で確定 )"
    rows="3"
    bind:value={appStore.commitMsg}
    onkeydown={onKeyDown}
  ></textarea>
  <div class="flex items-center gap-2">
    <span class="flex-1 text-[11px] text-fg-dim truncate">{appStore.status}</span>
    <button
      type="button"
      aria-label="claudeでコミットメッセージを生成"
      title="Claude Code で生成"
      class="w-7 h-7 flex items-center justify-center rounded border border-[var(--color-border)] text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg-softer)] disabled:opacity-40 disabled:cursor-not-allowed"
      disabled={appStore.generating || appStore.stagedFiles.length === 0}
      onclick={() => appStore.generateCommitMessage()}>
      {#if appStore.generating}
        <span class="inline-block w-3 h-3 rounded-full border-2 border-current border-t-transparent animate-spin"></span>
      {:else}
        <Icon name="sparkles" size={14} />
      {/if}
    </button>
    <button
      type="button"
      class="px-3 py-1 rounded bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white text-[12px] font-semibold disabled:opacity-50"
      disabled={!appStore.commitMsg.trim() || appStore.stagedFiles.length === 0}
      onclick={() => appStore.commit()}>コミット</button>
  </div>
</div>
