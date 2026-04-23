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
} from '../../../wailsjs/go/main/App'

class AppStore {
  info = $state<RepoInfo>({ root: '', branch: '' })
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
      this.status = '選択行なし'
      return
    }
    await this.guard(() => StageLines(this.selectedPath, hunks as any))
  }

  async unstageSelectedLines() {
    if (!this.selectedPath || !this.selectedStaged) return
    const hunks = this.buildSelectionPatch()
    if (hunks.length === 0) {
      this.status = '選択行なし'
      return
    }
    await this.guard(() => UnstageLines(this.selectedPath, hunks as any))
  }

  async commit() {
    const m = this.commitMsg.trim()
    if (!m) {
      this.status = 'コミットメッセージを入力してね'
      return
    }
    try {
      await Commit(m)
      this.commitMsg = ''
      this.status = 'コミット完了'
      await this.refresh()
    } catch (e: any) {
      this.status = `commit error: ${e?.message ?? e}`
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
