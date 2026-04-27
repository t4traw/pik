<script lang="ts">
  import { appStore } from '../stores/app.svelte'
  import { t } from '../i18n/index.svelte'
  import Icon from './Icon.svelte'

  type Row = { keys: string[]; descKey: Parameters<typeof t>[0] }
  type Section = { titleKey: Parameters<typeof t>[0]; rows: Row[] }

  const sections: Section[] = [
    {
      titleKey: 'shortcuts.section.navigation',
      rows: [
        { keys: ['↑', 'Ctrl', 'P'], descKey: 'shortcuts.row.prevFile' },
        { keys: ['↓', 'Ctrl', 'N'], descKey: 'shortcuts.row.nextFile' },
        { keys: ['←', 'Ctrl', 'B'], descKey: 'shortcuts.row.toStaged' },
        { keys: ['→', 'Ctrl', 'F'], descKey: 'shortcuts.row.toUnstaged' },
        { keys: ['Tab'], descKey: 'shortcuts.row.cyclePanels' },
        { keys: ['Shift', 'Tab'], descKey: 'shortcuts.row.cyclePanelsReverse' },
      ],
    },
    {
      titleKey: 'shortcuts.section.fileOps',
      rows: [
        { keys: ['Space'], descKey: 'shortcuts.row.toggleStage' },
        { keys: ['D'], descKey: 'shortcuts.row.discardOrDelete' },
      ],
    },
    {
      titleKey: 'shortcuts.section.commit',
      rows: [
        { keys: ['⌘', 'Enter'], descKey: 'shortcuts.row.commit' },
        { keys: ['⌘', 'Shift', 'Enter'], descKey: 'shortcuts.row.focusAndGenerate' },
      ],
    },
    {
      titleKey: 'shortcuts.section.history',
      rows: [
        { keys: ['⌘', 'Z'], descKey: 'shortcuts.row.undo' },
        { keys: ['⌘', 'Shift', 'Z'], descKey: 'shortcuts.row.redo' },
      ],
    },
  ]

  function close() {
    appStore.shortcutsOpen = false
  }

  function onKeyDown(e: KeyboardEvent) {
    if (appStore.shortcutsOpen && e.key === 'Escape') close()
  }
</script>

<svelte:window onkeydown={onKeyDown} />

{#if appStore.shortcutsOpen}
  <div
    class="fixed inset-0 z-40 bg-black/60"
    role="button"
    tabindex="-1"
    aria-label={t('shortcuts.closeAria')}
    onclick={close}
    onkeydown={(e) => e.key === 'Enter' && close()}
  ></div>

  <div
    class="fixed inset-0 z-50 flex items-center justify-center pointer-events-none"
    role="dialog"
    aria-modal="true"
    aria-label={t('shortcuts.title')}
  >
    <div
      class="pointer-events-auto w-[460px] max-h-[80vh] flex flex-col rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-soft)] shadow-xl"
    >
      <div
        class="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)]"
      >
        <span class="text-sm font-semibold">{t('shortcuts.title')}</span>
        <button
          type="button"
          aria-label={t('settings.close')}
          class="w-6 h-6 flex items-center justify-center rounded text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg-softer)]"
          onclick={close}
        >
          <Icon name="close" size={14} />
        </button>
      </div>

      <div class="p-4 space-y-4 overflow-y-auto">
        {#each sections as sec}
          <div>
            <div
              class="text-[11px] font-semibold tracking-wider text-[var(--color-fg-muted)] mb-2"
            >
              {t(sec.titleKey).toUpperCase()}
            </div>
            <ul class="space-y-1.5">
              {#each sec.rows as row}
                <li class="flex items-center gap-3 text-[13px]">
                  <span class="flex items-center gap-1 shrink-0">
                    {#each row.keys as key, i}
                      {#if i > 0}
                        <span class="text-[var(--color-fg-dim)] text-[11px]">+</span>
                      {/if}
                      <kbd
                        class="inline-flex items-center justify-center min-w-[24px] h-[22px] px-1.5 rounded border border-[var(--color-border)] bg-[var(--color-bg)] text-[var(--color-fg)] text-[11px] font-mono tabular-nums shadow-[inset_0_-1px_0_rgba(0,0,0,0.3)]"
                      >
                        {key}
                      </kbd>
                    {/each}
                  </span>
                  <span class="text-[var(--color-fg)] flex-1">{t(row.descKey)}</span>
                </li>
              {/each}
            </ul>
          </div>
        {/each}

        <div class="text-[11px] text-[var(--color-fg-dim)] pt-2 border-t border-[var(--color-border)]">
          {t('shortcuts.note')}
        </div>
      </div>
    </div>
  </div>
{/if}
