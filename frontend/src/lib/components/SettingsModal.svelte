<script lang="ts">
  import { appStore } from '../stores/app.svelte'
  import { t, type LocalePref } from '../i18n/index.svelte'
  import Icon from './Icon.svelte'

  const MIN = 8
  const MAX = 24

  function close() {
    appStore.settingsOpen = false
  }

  function onKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape') close()
  }

  async function setFontSize(n: number) {
    const clamped = Math.max(MIN, Math.min(MAX, Math.round(n)))
    await appStore.saveSettings({ ...appStore.settings, fontSize: clamped })
  }

  async function setLanguage(v: string) {
    const lang: LocalePref = v === 'en' || v === 'ja' ? v : ''
    await appStore.saveSettings({ ...appStore.settings, language: lang })
  }
</script>

<svelte:window onkeydown={onKeyDown} />

{#if appStore.settingsOpen}
  <!-- Backdrop — click to dismiss. -->
  <div
    class="fixed inset-0 z-40 bg-black/60"
    role="button"
    tabindex="-1"
    aria-label={t('settings.closeAria')}
    onclick={close}
    onkeydown={(e) => e.key === 'Enter' && close()}
  ></div>

  <div
    class="fixed inset-0 z-50 flex items-center justify-center pointer-events-none"
    role="dialog"
    aria-modal="true"
    aria-label={t('settings.title')}
  >
    <div class="pointer-events-auto w-[380px] rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-soft)] shadow-xl">
      <div class="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)]">
        <span class="text-sm font-semibold">{t('settings.title')}</span>
        <button
          type="button"
          aria-label={t('settings.close')}
          class="w-6 h-6 flex items-center justify-center rounded text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg-softer)]"
          onclick={close}>
          <Icon name="close" size={14} />
        </button>
      </div>

      <div class="p-4 space-y-4">
        <label class="block">
          <span class="block text-[12px] text-[var(--color-fg-muted)] mb-1">
            {t('settings.fontSize', { min: MIN, max: MAX })}
          </span>
          <div class="flex items-center gap-3">
            <input
              type="range"
              min={MIN}
              max={MAX}
              step="1"
              value={appStore.settings.fontSize}
              oninput={(e) => setFontSize(+e.currentTarget.value)}
              class="flex-1"
            />
            <input
              type="number"
              min={MIN}
              max={MAX}
              value={appStore.settings.fontSize}
              onchange={(e) => setFontSize(+e.currentTarget.value)}
              class="w-16 bg-[var(--color-bg)] border border-[var(--color-border)] rounded px-2 py-1 text-[12px] tabular-nums"
            />
          </div>
        </label>

        <label class="block">
          <span class="block text-[12px] text-[var(--color-fg-muted)] mb-1">
            {t('settings.language')}
          </span>
          <select
            value={appStore.settings.language}
            onchange={(e) => setLanguage(e.currentTarget.value)}
            class="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded px-2 py-1 text-[12px]"
          >
            <option value="">{t('settings.languageAuto')}</option>
            <option value="en">{t('settings.languageEn')}</option>
            <option value="ja">{t('settings.languageJa')}</option>
          </select>
        </label>
      </div>
    </div>
  </div>
{/if}
