# pik

日本語 | [English](./README.md)

**行単位のステージング** ができる軽量 Git GUI (macOS)。[Wails](https://wails.io) (Go + Svelte 5) 製。

## インストール

```sh
brew install t4traw/pik/pik
```

任意のリポジトリで起動:

```sh
pik          # カレントディレクトリ
pik ~/repo   # パス指定
```

## 機能

- ファイル / ハンク / **行単位** のステージ・アンステージ
- クリックで行選択できる diff ビュー (Shift クリックで範囲選択)
- UI からそのまま commit (⌘↩ で確定)
- ウィンドウフォーカス時に自動リフレッシュ — 外部エディタでの変更もすぐ反映
- 全操作に対する Undo / Redo (⌘Z / ⌘⇧Z)
- 設定パネルから diff のフォントサイズを変更
- **AI コミットメッセージ生成** — ローカルの `claude` CLI を叩いて、ステージ済み差分から 1 クリックでコミットメッセージを生成。リポジトリの `CLAUDE.md` の規約も自動で反映される。

## キーボードショートカット

| ショートカット | 動作 |
|---|---|
| ⌘↩ / Ctrl+↩ | コミット |
| ⌘Z / Ctrl+Z | 直前の操作を取り消し |
| ⌘⇧Z / Ctrl+⇧Z | やり直し |
| Shift + クリック | ハンク内の行を範囲選択 |

## 開発

Go 1.23+、[Bun](https://bun.sh)、[Wails CLI](https://wails.io/docs/gettingstarted/installation) が必要。

```sh
make dev       # ホットリロード開発サーバー
make build     # build/bin/pik.app を生成
make install   # $HOME/.local/bin/pik に shim を配置
```

## リリース

[release-please](https://github.com/googleapis/release-please) で自動化。release PR を merge すると tag push が走り、universal macOS ビルド・GitHub Release・Homebrew tap の formula bump まで全部自動。

## ライセンス

MIT
