import type { FileStatus, FileDiff, RepoInfo, PatchHunk } from '../types'
import {
  Info,
  Status,
  Diff,
  Stage,
  Unstage,
  StageAll,
  UnstageAll,
  Discard,
  StageLines,
  UnstageLines,
  Commit,
  Undo,
  Redo,
  UndoState,
  GetSettings,
  UpdateSettings,
  GenerateCommitMessage,
  DetectLocale,
  Sync,
} from '../../../wailsjs/go/main/App'
import { i18n, t, type Locale, type LocalePref } from '../i18n/index.svelte'

type UndoInfo = { canUndo: boolean; canRedo: boolean; undoDesc: string; redoDesc: string }
export type Settings = { fontSize: number; language: LocalePref }

class AppStore {
  info = $state<RepoInfo>({ root: '', branch: '', ahead: 0, behind: 0, hasUpstream: false })
  syncing = $state<boolean>(false)
  files = $state<FileStatus[]>([])

  selectedPath = $state<string>('')
  selectedStaged = $state<boolean>(false)
  diffFiles = $state<FileDiff[]>([])
  loading = $state<boolean>(false)

  // For line-level selection in the diff.
  // Key format: `${hunkIdx}:${lineIdx}` (per selected file)
  selectedLines = $state<Set<string>>(new Set())

  commitMsg = $state<string>('')
  status = $state<string>('')

  undo_ = $state<UndoInfo>({ canUndo: false, canRedo: false, undoDesc: '', redoDesc: '' })

  settings = $state<Settings>({ fontSize: 12, language: '' })
  settingsOpen = $state<boolean>(false)
  shortcutsOpen = $state<boolean>(false)

  generating = $state<boolean>(false)

  get selectedFile(): FileStatus | undefined {
    return this.files.find(
      (f) =>
        f.Path === this.selectedPath &&
        (this.selectedStaged ? f.Staged : f.Unstaged || f.Untracked),
    )
  }

  get stagedFiles(): FileStatus[] {
    return this.files.filter((f) => f.Staged)
  }

  get unstagedFiles(): FileStatus[] {
    return this.files.filter((f) => f.Unstaged || f.Untracked)
  }

  async refresh() {
    try {
      this.info = await Info()
      this.files = (await Status()) ?? []
      this.undo_ = await UndoState()
    } catch (e: any) {
      this.status = `status error: ${e?.message ?? e}`
      return
    }
    // keep current selection if still present, else clear
    const f = this.files.find((x) => x.Path === this.selectedPath)
    if (f) {
      const stillStaged = this.selectedStaged && f.Staged
      const stillUnstaged = !this.selectedStaged && (f.Unstaged || f.Untracked)
      if (stillStaged || stillUnstaged) {
        await this.loadDiff(f, this.selectedStaged)
        return
      }
    }
    this.selectedPath = ''
    this.selectedStaged = false
    this.diffFiles = []
    this.selectedLines = new Set()
  }

  async selectFile(file: FileStatus, staged: boolean) {
    if (this.selectedPath === file.Path && this.selectedStaged === staged) {
      return
    }
    this.selectedPath = file.Path
    this.selectedStaged = staged
    this.selectedLines = new Set()
    await this.loadDiff(file, staged)
  }

  async moveSelection(direction: 'up' | 'down' | 'left' | 'right') {
    const staged = this.stagedFiles
    const unstaged = this.unstagedFiles
    const flat: Array<{ file: FileStatus; staged: boolean }> = [
      ...staged.map((f) => ({ file: f, staged: true })),
      ...unstaged.map((f) => ({ file: f, staged: false })),
    ]
    if (flat.length === 0) return

    const idx = flat.findIndex(
      (x) => x.file.Path === this.selectedPath && x.staged === this.selectedStaged,
    )

    if (direction === 'up' || direction === 'down') {
      const delta = direction === 'down' ? 1 : -1
      const next =
        idx < 0
          ? direction === 'down'
            ? 0
            : flat.length - 1
          : Math.max(0, Math.min(flat.length - 1, idx + delta))
      const t = flat[next]
      await this.selectFile(t.file, t.staged)
      return
    }

    // left: jump to the staged side. right: jump to the unstaged side.
    // Prefer the twin (same path) on the other side if it exists, else fall back
    // to the first file in the target section.
    const wantStaged = direction === 'left'
    const cur = idx >= 0 ? flat[idx] : undefined
    if (cur && cur.staged !== wantStaged) {
      const twin = flat.find((x) => x.file.Path === cur.file.Path && x.staged === wantStaged)
      if (twin) {
        await this.selectFile(twin.file, twin.staged)
        return
      }
    }
    const targetList = wantStaged ? staged : unstaged
    if (targetList.length > 0) {
      await this.selectFile(targetList[0], wantStaged)
    }
  }

  async loadDiff(file: FileStatus, staged: boolean) {
    this.loading = true
    try {
      const r = await Diff(file.Path, staged, file.Untracked)
      this.diffFiles = r?.files ?? []
    } catch (e: any) {
      this.status = `diff error: ${e?.message ?? e}`
      this.diffFiles = []
    } finally {
      this.loading = false
    }
  }

  toggleLine(hunkIdx: number, lineIdx: number) {
    const k = `${hunkIdx}:${lineIdx}`
    const s = new Set(this.selectedLines)
    if (s.has(k)) s.delete(k)
    else s.add(k)
    this.selectedLines = s
  }

  selectAllInHunk(hunkIdx: number) {
    const h = this.diffFiles[0]?.hunks[hunkIdx]
    if (!h) return
    const s = new Set(this.selectedLines)
    h.lines.forEach((l, i) => {
      if (l.op !== 'context') s.add(`${hunkIdx}:${i}`)
    })
    this.selectedLines = s
  }

  clearLineSelection() {
    this.selectedLines = new Set()
  }

  hasLineSelection(): boolean {
    return this.selectedLines.size > 0
  }

  async toggleStageSelected() {
    const f = this.selectedFile
    if (!f) return
    if (this.selectedStaged) await this.unstage(f.Path)
    else await this.stage(f.Path)
  }

  async discardSelected() {
    const f = this.selectedFile
    if (!f || this.selectedStaged) return
    const msg = f.Untracked
      ? t('fileList.confirmDeleteUntracked', { path: f.Path })
      : t('fileList.confirmDiscardChanges', { path: f.Path })
    if (confirm(msg)) await this.discard(f.Path, f.Untracked)
  }

  async stage(path: string) {
    await this.guard(() => Stage(path))
  }
  async unstage(path: string) {
    await this.guard(() => Unstage(path))
  }
  async stageAll() {
    await this.guard(() => StageAll())
  }
  async unstageAll() {
    await this.guard(() => UnstageAll())
  }
  async discard(path: string, untracked: boolean) {
    await this.guard(() => Discard(path, untracked))
  }

  /** Build the list of selected lines into PatchHunk[] for staging. */
  private buildSelectionPatch(): PatchHunk[] {
    const file = this.diffFiles[0]
    if (!file) return []
    const out: PatchHunk[] = []
    file.hunks.forEach((h, hi) => {
      const lines = h.lines.map((l, li) => ({
        op: l.op,
        text: l.text,
        selected: this.selectedLines.has(`${hi}:${li}`),
      }))
      // Only include hunks that have at least one selected change.
      const hasChange = lines.some((l) => l.op !== 'context' && l.selected)
      if (hasChange) {
        out.push({ oldStart: h.oldStart, newStart: h.newStart, lines })
      }
    })
    return out
  }

  async stageSelectedLines() {
    if (!this.selectedPath || this.selectedStaged) return
    const hunks = this.buildSelectionPatch()
    if (hunks.length === 0) {
      this.status = t('status.noLineSelected')
      return
    }
    await this.guard(() => StageLines(this.selectedPath, hunks as any))
  }

  async unstageSelectedLines() {
    if (!this.selectedPath || !this.selectedStaged) return
    const hunks = this.buildSelectionPatch()
    if (hunks.length === 0) {
      this.status = t('status.noLineSelected')
      return
    }
    await this.guard(() => UnstageLines(this.selectedPath, hunks as any))
  }

  async commit() {
    const m = this.commitMsg.trim()
    if (!m) {
      this.status = t('status.commitMessageRequired')
      return
    }
    try {
      await Commit(m)
      this.commitMsg = ''
      this.status = t('status.commitDone')
      await this.refresh()
    } catch (e: any) {
      this.status = `commit error: ${e?.message ?? e}`
    }
  }

  async undo() {
    try {
      const desc = await Undo()
      if (!desc) {
        this.status = t('status.nothingToUndo')
        return
      }
      this.status = t('status.undid', { desc })
      await this.refresh()
    } catch (e: any) {
      this.status = `undo error: ${e?.message ?? e}`
    }
  }

  async redo() {
    try {
      const desc = await Redo()
      if (!desc) {
        this.status = t('status.nothingToRedo')
        return
      }
      this.status = t('status.redid', { desc })
      await this.refresh()
    } catch (e: any) {
      this.status = `redo error: ${e?.message ?? e}`
    }
  }

  private async resolveLocale(pref: LocalePref): Promise<Locale> {
    if (pref === 'en' || pref === 'ja') return pref
    try {
      const detected = await DetectLocale()
      return detected === 'ja' ? 'ja' : 'en'
    } catch {
      return 'en'
    }
  }

  async loadSettings() {
    try {
      this.settings = ((await GetSettings()) as unknown as Settings) ?? this.settings
      i18n.set(await this.resolveLocale(this.settings.language))
    } catch (e: any) {
      this.status = `settings load error: ${e?.message ?? e}`
    }
  }

  async saveSettings(next: Settings) {
    try {
      this.settings = ((await UpdateSettings(next as any)) as unknown as Settings) ?? next
      i18n.set(await this.resolveLocale(this.settings.language))
    } catch (e: any) {
      this.status = `settings save error: ${e?.message ?? e}`
    }
  }

  async sync() {
    if (this.syncing) return
    this.syncing = true
    this.status = t('status.syncing')
    try {
      const summary = await Sync()
      this.status = t('status.syncDone', { summary })
      await this.refresh()
    } catch (e: any) {
      this.status = `${e?.message ?? e}`
    } finally {
      this.syncing = false
    }
  }

  async generateCommitMessage() {
    if (this.generating) return
    if (this.stagedFiles.length === 0) {
      this.status = t('status.noStagedChanges')
      return
    }
    this.generating = true
    this.status = t('status.generating')
    try {
      const msg = await GenerateCommitMessage()
      if (msg) {
        this.commitMsg = msg
        this.status = t('status.generated')
      }
    } catch (e: any) {
      this.status = `${e?.message ?? e}`
    } finally {
      this.generating = false
    }
  }

  private async guard(fn: () => Promise<void>) {
    try {
      await fn()
      await this.refresh()
    } catch (e: any) {
      this.status = `${e?.message ?? e}`
    }
  }
}

export const appStore = new AppStore()
