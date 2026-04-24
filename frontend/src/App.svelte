<script lang="ts">
  import { onMount } from 'svelte'
  import { appStore } from './lib/stores/app.svelte'
  import FileList from './lib/components/FileList.svelte'
  import DiffView from './lib/components/DiffView.svelte'
  import CommitBox from './lib/components/CommitBox.svelte'
  import Icon from './lib/components/Icon.svelte'
  import SettingsModal from './lib/components/SettingsModal.svelte'

  onMount(() => {
    appStore.refresh()
    appStore.loadSettings()
    const onFocus = () => appStore.refresh()
    const PANEL_IDS = ['pik-panel-files', 'pik-panel-diff', 'pik-panel-commit'] as const

    const focusCommit = () => {
      ;(document.getElementById('pik-panel-commit') as HTMLTextAreaElement | null)?.focus()
    }

    const cyclePanel = (shift: boolean) => {
      const active = document.activeElement
      let cur = PANEL_IDS.findIndex((id) => {
        const el = document.getElementById(id)
        return !!el && (el === active || el.contains(active))
      })
      if (cur === -1) cur = 0
      const delta = shift ? -1 : 1
      const next = (cur + delta + PANEL_IDS.length) % PANEL_IDS.length
      document.getElementById(PANEL_IDS[next])?.focus()
    }

    const isEditable = (t: EventTarget | null): boolean => {
      const el = t as HTMLElement | null
      return !!el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)
    }

    const onKey = (e: KeyboardEvent) => {
      if (e.isComposing) return
      const k = e.key.toLowerCase()
      const mod = e.metaKey || e.ctrlKey
      const editable = isEditable(e.target)

      // Undo / Redo (global).
      if (mod && k === 'z') {
        e.preventDefault()
        if (e.shiftKey) appStore.redo()
        else appStore.undo()
        return
      }

      // Cmd/Ctrl + Shift + Enter: focus commit box AND trigger Claude generation.
      if (mod && e.shiftKey && e.key === 'Enter') {
        e.preventDefault()
        focusCommit()
        appStore.generateCommitMessage()
        return
      }

      // Cmd/Ctrl + Enter: commit. Inside the textarea the component's own
      // handler already fires, so skip here to avoid a double-commit.
      if (mod && !e.shiftKey && e.key === 'Enter') {
        if (editable) return
        e.preventDefault()
        appStore.commit()
        return
      }

      // Tab / Shift+Tab: cycle between the three panels.
      if (e.key === 'Tab' && !mod && !e.altKey) {
        e.preventDefault()
        cyclePanel(e.shiftKey)
        return
      }

      // Remaining shortcuts all mutate files — skip while typing.
      if (editable) return

      // Space: toggle stage / unstage on the selected file.
      if (e.key === ' ' || e.code === 'Space') {
        e.preventDefault()
        appStore.toggleStageSelected()
        return
      }

      // D: discard the selected unstaged / untracked file.
      if (k === 'd' && !mod && !e.altKey && !e.shiftKey) {
        e.preventDefault()
        appStore.discardSelected()
        return
      }

      // Emacs-style bindings use the physical Control key only (not Cmd).
      const emacs = e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey
      let dir: 'up' | 'down' | 'left' | 'right' | null = null
      if (e.key === 'ArrowUp' || (emacs && k === 'p')) dir = 'up'
      else if (e.key === 'ArrowDown' || (emacs && k === 'n')) dir = 'down'
      else if (e.key === 'ArrowLeft' || (emacs && k === 'b')) dir = 'left'
      else if (e.key === 'ArrowRight' || (emacs && k === 'f')) dir = 'right'
      if (dir) {
        e.preventDefault()
        appStore.moveSelection(dir)
      }
    }
    window.addEventListener('focus', onFocus)
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('focus', onFocus)
      window.removeEventListener('keydown', onKey)
    }
  })
</script>

<div class="flex flex-col h-full" style="--pik-font-size: {appStore.settings.fontSize}px;">
  <!-- Title bar (drag region, matches macOS traffic-light layout) -->
  <div
    class="flex items-center gap-3 pr-2 bg-[var(--color-bg-soft)] border-b border-[var(--color-border)] text-[12px] select-none"
    style="--wails-draggable: drag; padding-left: 80px; height: 28px;"
  >
    <span class="flex items-center gap-1.5 text-sky-300 font-semibold shrink-0">
      <Icon name="branch" size={14} />
      <span>{appStore.info.branch || '—'}</span>
    </span>
    <span class="text-[var(--color-fg-muted)] truncate flex-1 min-w-0">{appStore.info.root}</span>
    <button
      type="button"
      aria-label="Refresh"
      class="shrink-0 w-7 h-7 flex items-center justify-center rounded text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg-softer)] transition-colors"
      style="--wails-draggable: no-drag;"
      onclick={() => appStore.refresh()}>
      <Icon name="refresh" size={15} />
    </button>
    <button
      type="button"
      aria-label="Settings"
      class="shrink-0 w-7 h-7 flex items-center justify-center rounded text-[var(--color-fg-muted)] hover:text-white hover:bg-[var(--color-bg-softer)] transition-colors"
      style="--wails-draggable: no-drag;"
      onclick={() => (appStore.settingsOpen = true)}>
      <Icon name="settings" size={15} />
    </button>
  </div>

  <SettingsModal />

  <!-- Main split -->
  <div class="flex-1 grid overflow-hidden" style="grid-template-columns: minmax(260px, 340px) 1fr;">
    <!-- Left column: file list + commit box -->
    <div class="flex flex-col border-r border-[var(--color-border)] overflow-hidden">
      <div
        id="pik-panel-files"
        tabindex="-1"
        class="flex-1 overflow-hidden outline-none focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-[var(--color-accent)]"
      >
        <FileList />
      </div>
      <CommitBox />
    </div>

    <!-- Right column: diff -->
    <div
      id="pik-panel-diff"
      tabindex="-1"
      class="flex flex-col overflow-hidden bg-[var(--color-bg)] outline-none focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-[var(--color-accent)]"
    >
      <DiffView />
    </div>
  </div>
</div>
