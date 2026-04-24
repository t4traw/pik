<script lang="ts">
  import { appStore } from '../stores/app.svelte'
  import Icon from './Icon.svelte'

  type Row = { keys: string[]; desc: string }
  type Section = { title: string; rows: Row[] }

  const sections: Section[] = [
    {
      title: 'ナビゲーション',
      rows: [
        { keys: ['↑', 'Ctrl', 'P'], desc: '前のファイル' },
        { keys: ['↓', 'Ctrl', 'N'], desc: '次のファイル' },
        { keys: ['←', 'Ctrl', 'B'], desc: 'ステージ側へ' },
        { keys: ['→', 'Ctrl', 'F'], desc: '未ステージ側へ' },
        { keys: ['Tab'], desc: 'パネル巡回 (ファイル → Diff → コミット欄)' },
        { keys: ['Shift', 'Tab'], desc: '逆方向にパネル巡回' },
      ],
    },
    {
      title: 'ファイル操作',
      rows: [
        { keys: ['Space'], desc: 'ステージ / アンステージをトグル' },
        { keys: ['D'], desc: '変更を破棄 / 未追跡ファイルを削除' },
      ],
    },
    {
      title: 'コミット',
      rows: [
        { keys: ['⌘', 'Enter'], desc: 'コミット' },
        { keys: ['⌘', 'Shift', 'Enter'], desc: 'コミット欄にフォーカス + Claude で生成' },
      ],
    },
    {
      title: '編集履歴',
      rows: [
        { keys: ['⌘', 'Z'], desc: '元に戻す' },
        { keys: ['⌘', 'Shift', 'Z'], desc: 'やり直し' },
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
    aria-label="ショートカット一覧を閉じる"
    onclick={close}
    onkeydown={(e) => e.key === 'Enter' && close()}
  ></div>

  <div
    class="fixed inset-0 z-50 flex items-center justify-center pointer-events-none"
    role="dialog"
    aria-modal="true"
    aria-label="キーボードショートカット"
  >
    <div
      class="pointer-events-auto w-[460px] max-h-[80vh] flex flex-col rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-soft)] shadow-xl"
    >
      <div
        class="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)]"
      >
        <span class="text-sm font-semibold">キーボードショートカット</span>
        <button
          type="button"
          aria-label="閉じる"
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
              {sec.title.toUpperCase()}
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
                  <span class="text-[var(--color-fg)] flex-1">{row.desc}</span>
                </li>
              {/each}
            </ul>
          </div>
        {/each}

        <div class="text-[11px] text-[var(--color-fg-dim)] pt-2 border-t border-[var(--color-border)]">
          テキスト入力中は Space / D / 矢印 はタイピングが優先されます。
        </div>
      </div>
    </div>
  </div>
{/if}
