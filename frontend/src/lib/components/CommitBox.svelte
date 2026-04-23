<script lang="ts">
  import { appStore } from '../stores/app.svelte'

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
      class="px-3 py-1 rounded bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white text-[12px] font-semibold disabled:opacity-50"
      disabled={!appStore.commitMsg.trim() || appStore.stagedFiles.length === 0}
      onclick={() => appStore.commit()}>コミット</button>
  </div>
</div>
