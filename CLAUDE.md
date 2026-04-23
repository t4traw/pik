# pik — プロジェクト作業ルール

このレポは **release-please + Conventional Commits** で自動リリースしている。
コミットメッセージを規定の prefix で書かないとバージョン bump されないので、
AI が commit を作る場合は **必ず以下の形式に従う**。

## コミットメッセージ形式 (必須)

```
<prefix>: <短い要約>

<オプションの詳細本文>
```

- 要約は現在形・命令形ぽく書く (英語でも日本語でも可)
- 本文は必要なときだけ。空行で区切る
- 破壊的変更がある場合は prefix の後に `!` を付ける or 本文に `BREAKING CHANGE:` を含める

### 使う prefix と効果

| prefix | 意味 | バージョン bump |
|---|---|---|
| `feat:` | ユーザー向けの新機能 | minor (0.1.0 → 0.2.0) |
| `fix:` | バグ修正 | patch (0.1.0 → 0.1.1) |
| `perf:` | パフォーマンス改善 | patch |
| `refactor:` | 挙動変わらないリファクタ | リリースに含めない |
| `docs:` | ドキュメントのみ | リリースに含めない |
| `test:` | テスト追加・修正 | リリースに含めない |
| `chore:` | ビルド・CI・雑務 | リリースに含めない |
| `build:` | ビルドシステム変更 | リリースに含めない |
| `ci:` | GitHub Actions など CI 設定 | リリースに含めない |
| `style:` | フォーマットのみ | リリースに含めない |
| `feat!:` / `fix!:` / `BREAKING CHANGE:` | 破壊的変更 | 1.0.0 未満は minor bump |

### 例

```
feat: 行単位のステージングに対応

選択した行だけを index に入れるためのpatch構築を追加。
UIから diff の行をクリックして `選択をステージ` できる。
```

```
fix: 日本語IME確定時の余分な改行を抑止

WebView 側のtextarea で Enter が重複で発火していた問題を修正。
```

```
chore: deps を更新
```

```
ci: release-please ワークフローを追加
```

## やってはいけないこと

- prefix なし / 独自 prefix (例: `update`, `change`, `WIP`, `UI調整`) → release-please が無視するので **絶対使わない**
- 1コミットに複数の prefix 的な変更を混ぜない。分けて commit する
- リリース対象外の変更 (ビルド設定など) は `chore:` / `ci:` にして、`feat:` に混ぜない

## リリースフロー (参考)

1. `feat:` / `fix:` で commit → push
2. `release-please` が自動で release PR を開く
3. PR を merge → tag 作成 & GitHub Release + tarball + Homebrew tap 更新、全部自動

詳細は `dist-templates/README.md`。

## その他

- Wails 製アプリ (Go バックエンド + Svelte 5 frontend)
- `make dev` でホットリロード開発、`make install` でローカルの PATH に配置
- 起動コマンド: `pik [repo-path]` (cwd がデフォルト)
