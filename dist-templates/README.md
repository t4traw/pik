# Release & Distribution Guide

## 1. 初回セットアップ (一度だけ)

### Homebrew tap リポジトリを作る

GitHub で **`t4traw/homebrew-pik`** という名前のパブリックレポジトリを作る。
(`homebrew-` プレフィックスは Homebrew の規約で必須。)

```
homebrew-pik/
├── Formula/
│   └── pik.rb        ← このレポの dist-templates/pik.rb をコピー
└── README.md
```

> ⚠️ `Formula/` ディレクトリが重要 (Cask ではない)。
> Formula にすると `brew install t4traw/pik/pik` の短縮形が使える。

### ユーザーのインストール手順

```sh
brew install t4traw/pik/pik
# もしくは tap 分離:
brew tap t4traw/pik
brew install pik
```

これで `pik` コマンドが PATH に入り、GUI アプリ本体は `$(brew --prefix)/opt/pik/pik.app` に配置される。

## 2. PAT を一度だけ設定 (Homebrew tap 自動更新用)

1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
   → **Generate new token (classic)**
2. Scope は `repo` だけでOK。Expiration は長めに (1 year など)
3. 発行されたトークンをコピー
4. `t4traw/pik` リポの Settings → Secrets and variables → Actions →
   **New repository secret**
   - Name: `TAP_GITHUB_TOKEN`
   - Value: コピーしたトークン

これで Actions から `t4traw/homebrew-pik` に push できるようになる。

## 3. 日々のリリース手順

### 3-1. コミットは conventional commits で

| prefix | 意味 | バージョン影響 |
|---|---|---|
| `feat: ...` | 新機能 | minor up (0.1.0 → 0.2.0) |
| `fix: ...` | バグ修正 | patch up (0.1.0 → 0.1.1) |
| `feat!: ...` or `fix!: ...` | 破壊的変更 | 1.0.0 未満は minor up |
| `chore: ...` / `docs: ...` / `refactor: ...` / `test: ...` | リリースに含めない | なし |

例:

```sh
git commit -m "feat: 複数コミットへのドラッグ分類に対応"
git commit -m "fix: 日本語入力時の改行を抑止"
```

### 3-2. push する

```sh
git push
```

### 3-3. release-please が自動で release PR を開く

`.github/workflows/release-please.yml` が main への push を watch して、
次回分のバージョンと CHANGELOG.md 更新案を含む PR を開く:

- タイトル: `chore: release 0.2.0`
- 中身: CHANGELOG.md に feat/fix のまとめ + `.release-please-manifest.json` のバージョン bump

### 3-4. PR を merge する

この PR を merge するだけ。以降すべて自動:

1. release-please がタグ (`v0.2.0`) + GitHub Release を作成
2. タグ push で `release.yml` が発火
3. macOS runner で universal `.app` ビルド
4. tarball を先ほどの Release に添付
5. SHA-256 計算 → `t4traw/homebrew-pik/Formula/pik.rb` 自動更新

ユーザー側:

```sh
brew install t4traw/pik/pik     # 初回
brew upgrade pik                # 2回目以降
```

### ローカルで検証ビルドする場合

```sh
make release-build   # universal .app 作成
make release-tar     # dist/pik-<version>-darwin-universal.tar.gz + SHA表示
```

## 4. コード署名について

Wails の **ad-hoc self-sign** のみ。Homebrew 経由なら `brew install` 時に
quarantine attribute を自動で外してくれる (詳しくは `brew install --help`) ので
Gatekeeper 警告は基本出ない。

第三者サイトから直接 tarball ダウンロード → 展開 → 起動だと `"pik" cannot be
opened because it is from an unidentified developer` が出る。その場合は右クリック
→ Open で初回バイパス。

公式的に静かに動かすなら Apple Developer ID signing + notarization ($99/年) が
必要。tap を主動線にしておけば当面は不要。
