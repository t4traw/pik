// Mirrors Go types in internal/git. Kept manually so the frontend can stay
// typed even before the wailsjs bindings get regenerated.

export type LineOp = 'context' | 'add' | 'remove'

export interface FileStatus {
  Path: string
  IndexStatus: number
  WorkStatus: number
  Staged: boolean
  Unstaged: boolean
  Untracked: boolean
  Conflicted: boolean
}

export interface DiffLine {
  op: LineOp
  text: string
  oldLineNo: number
  newLineNo: number
}

export interface Hunk {
  oldStart: number
  newStart: number
  header: string
  lines: DiffLine[]
}

export interface FileDiff {
  oldPath: string
  newPath: string
  hunks: Hunk[]
  binary: boolean
  preamble: string[]
}

export interface DiffResult {
  files: FileDiff[]
  raw: string
}

export interface RepoInfo {
  root: string
  branch: string
  ahead: number
  behind: number
  hasUpstream: boolean
}

// For line-level staging.
export interface PatchLine {
  op: LineOp
  text: string
  selected: boolean
}

export interface PatchHunk {
  oldStart: number
  newStart: number
  lines: PatchLine[]
}
