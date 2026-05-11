<script lang="ts">
  import { appStore } from '../stores/app.svelte'
  import { t } from '../i18n/index.svelte'

  let okBtn = $state<HTMLButtonElement | null>(null)

  function cancel() {
    appStore.resolveConfirm(false)
  }

  function ok() {
    appStore.resolveConfirm(true)
  }

  function onKeyDown(e: KeyboardEvent) {
    if (!appStore.confirmOpen) return
    if (e.key === 'Escape') {
      e.preventDefault()
      e.stopPropagation()
      cancel()
    } else if (e.key === 'Enter') {
      e.preventDefault()
      e.stopPropagation()
      ok()
    }
  }

  // WKWebView ignores autofocus on dynamically-mounted elements, so focus the
  // OK button manually once the modal opens — Enter then works without Tab.
  $effect(() => {
    if (appStore.confirmOpen && okBtn) okBtn.focus()
  })
</script>

<svelte:window onkeydown={onKeyDown} />

{#if appStore.confirmOpen}
  <div
    class="fixed inset-0 z-[60] bg-black/60"
    role="button"
    tabindex="-1"
    aria-label={t('confirm.closeAria')}
    onclick={cancel}
    onkeydown={(e) => e.key === 'Enter' && cancel()}
  ></div>

  <div
    class="fixed inset-0 z-[70] flex items-center justify-center pointer-events-none"
    role="dialog"
    aria-modal="true"
  >
    <div class="pointer-events-auto w-[420px] max-w-[90vw] rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-soft)] shadow-xl">
      <div class="px-4 py-4 whitespace-pre-line break-words text-[13px] text-[var(--color-fg)]">
        {appStore.confirmMessage}
      </div>
      <div class="flex justify-end gap-2 px-3 pb-3">
        <button
          type="button"
          class="h-7 px-3 rounded text-[12px] text-[var(--color-fg)] bg-[var(--color-bg)] border border-[var(--color-border)] hover:bg-[var(--color-bg-softer)]"
          onclick={cancel}
        >
          {t('confirm.cancel')}
        </button>
        <button
          type="button"
          bind:this={okBtn}
          class="h-7 px-3 rounded text-[12px] text-white {appStore.confirmDanger ? 'bg-rose-700 hover:bg-rose-600' : 'bg-[var(--color-accent)] hover:brightness-110'}"
          onclick={ok}
        >
          {appStore.confirmLabel || t('confirm.ok')}
        </button>
      </div>
    </div>
  </div>
{/if}
