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

## 2. リリースを作るとき

### 本番リリース (推奨: GitHub Actions)

1. タグを切って push

   ```sh
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. `.github/workflows/release.yml` が自動で:
   - macOS runner で universal `.app` をビルド
   - `dist/pik-v0.1.0-darwin-universal.tar.gz` を作る
   - GitHub Release を作成、tarball を添付
   - SHA-256 を release notes に書き込む

3. Release notes に出た SHA-256 をコピー

4. **`t4traw/homebrew-pik` リポの `Formula/pik.rb`** を編集:
   - `version` を `"0.1.0"` に
   - `sha256` を先ほどのハッシュに差し替え
   - commit & push

5. ユーザーは `brew upgrade pik` で新版取得

### ローカルで検証ビルドする場合

```sh
make release-build   # universal .app 作成
make release-tar     # dist/pik-<version>-darwin-universal.tar.gz + SHA表示
```

## 3. Homebrew tap の自動更新 (optional)

`t4traw/homebrew-pik` 側に action を置いておくと、pik の release を watch して自動で formula を更新できる。

```yaml
# .github/workflows/bump.yml (in homebrew-pik repo)
name: bump-on-release
on:
  repository_dispatch:
    types: [pik-release]
  workflow_dispatch:
    inputs:
      version:
        required: true
      sha256:
        required: true

jobs:
  bump:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Update formula
        run: |
          V="${{ inputs.version }}"
          SHA="${{ inputs.sha256 }}"
          sed -i "s/version \".*\"/version \"${V#v}\"/" Formula/pik.rb
          sed -i "s/sha256 \".*\"/sha256 \"$SHA\"/" Formula/pik.rb
      - uses: stefanzweifel/git-auto-commit-action@v5
        with:
          commit_message: "pik ${{ inputs.version }}"
```

そして pik リポ側の `release.yml` で `repository_dispatch` を飛ばすステップを足せば
完全自動化。

## 4. コード署名について

Wails の **ad-hoc self-sign** のみ。Homebrew 経由なら `brew install` 時に
quarantine attribute を自動で外してくれる (詳しくは `brew install --help`) ので
Gatekeeper 警告は基本出ない。

第三者サイトから直接 tarball ダウンロード → 展開 → 起動だと `"pik" cannot be
opened because it is from an unidentified developer` が出る。その場合は右クリック
→ Open で初回バイパス。

公式的に静かに動かすなら Apple Developer ID signing + notarization ($99/年) が
必要。tap を主動線にしておけば当面は不要。
