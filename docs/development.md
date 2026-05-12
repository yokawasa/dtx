# dtx 開発者向けマニュアル

このドキュメントは、dtx本体を開発・テストする内部開発者向けの手順をまとめる。

利用者向けのインストール・基本利用手順は `README.md` を参照する。

## 前提

### 必須

* Go 1.21 以上
* dotenvx CLI

dotenvx CLIの存在確認：

```bash
command -v dotenvx
```

未インストールの場合：

```bash
npm install -g @dotenvx/dotenvx
```

## 基本の開発ループ

通常の変更後は、以下の順で確認する。

```bash
gofmt -w cmd internal
go test ./...
go build -o /private/tmp/dtx ./cmd/dtx
```

注意：

* ルートディレクトリに `dtx` バイナリを生成しない
* ビルド成果物は `/private/tmp/dtx` や `/tmp/dtx` へ出力する
* 実ユーザーの `~/.dtx` を汚さないため、手動確認では `DTX_HOME` を使う

## テスト

全テストを実行する。

```bash
go test ./...
```

特定パッケージだけ実行する。

```bash
go test ./internal/cli
go test ./internal/core
```

詳細ログを出す。

```bash
go test -v ./...
```

## ビルド

開発確認用のバイナリを作成する。

```bash
go build -o /private/tmp/dtx ./cmd/dtx
```

バージョン・ヘルプの簡易確認：

```bash
/private/tmp/dtx --help
/private/tmp/dtx --version
```

## 隔離された手動確認

`DTX_HOME` を一時ディレクトリへ向けることで、実ユーザーの `~/.dtx` を変更せずに確認できる。

```bash
tmp_home=$(mktemp -d /private/tmp/dtx-home.XXXXXX)
DTX_HOME="$tmp_home" /private/tmp/dtx ls
```

## 実dotenvx込みのE2E確認

`edit` から `run` までを、実際のdotenvx CLIを使って確認する。

```bash
go build -o /private/tmp/dtx ./cmd/dtx

tmp_home=$(mktemp -d /private/tmp/dtx-home.XXXXXX)
editor=$(mktemp /private/tmp/dtx-editor.XXXXXX)

cat > "$editor" <<'SH'
#!/bin/sh
printf 'HELLO=dev\n' > "$1"
SH
chmod +x "$editor"

DTX_HOME="$tmp_home" VISUAL="$editor" /private/tmp/dtx edit dev
DTX_HOME="$tmp_home" /private/tmp/dtx ls
DTX_HOME="$tmp_home" /private/tmp/dtx use dev
DTX_HOME="$tmp_home" /private/tmp/dtx current
DTX_HOME="$tmp_home" /private/tmp/dtx run dev -- sh -c 'printf "$HELLO\n"'
```

期待される出力：

```text
Edited env: dev
dev
Using env: dev
dev
Using env: dev
dev
```

## ディレクトリ構造

主要な実装箇所：

```text
cmd/dtx/main.go        CLIエントリポイント
internal/cli/          引数解析とコマンド振り分け
internal/command/      dtx各コマンドの実装
internal/core/         path、env名、権限、env store
internal/dotenvx/      dotenvx CLI adapter
internal/process/      サブプロセス実行
internal/apperr/       終了コードとエラー整形
```

## dotenvx連携の確認ポイント

dotenvx CLI呼び出しは `internal/dotenvx` に閉じ込める。

確認する観点：

* dtx本体がdotenvxの暗号形式を解釈していないこと
* `run` / `encrypt` / `decrypt` 相当の呼び出しがadapter経由であること
* 通常時は `--quiet` を使い、`--verbose` 指定時のみ詳細出力を許可すること
* envファイルは `envs/<env>.enc`、鍵ファイルは `keys/<env>` に保存されること

## 権限確認

手動確認後、必要に応じてファイル権限を確認する。

```bash
find "$tmp_home" -maxdepth 2 -type d -exec ls -ld {} \;
find "$tmp_home" -maxdepth 2 -type f -exec ls -l {} \;
```

期待値：

* dtx home、`envs/`、`keys/` は `700`
* `current`、envファイル、鍵ファイルは `600`

## よく使う確認コマンド

現在の差分：

```bash
git status --short
git diff
```

Goファイル一覧：

```bash
find cmd internal -type f | sort
```

dotenvxのCLI仕様確認：

```bash
dotenvx run --help
dotenvx encrypt --help
dotenvx decrypt --help
```

## リリース

リリースは GitHub Actions + GoReleaser で自動化されている。バージョンタグを push するだけで完結する。

### リリース手順

```bash
# 1. main が最新の状態であることを確認
git checkout main
git pull

# 2. 全テストが通ることを確認
go test ./...

# 3. バージョンタグを付けて push（これだけで自動リリースが走る）
git tag v0.1.0
git push origin v0.1.0
```

タグ push 後、`.github/workflows/release.yml` が起動し、GitHub Release に以下が自動生成される。

| 成果物 | 内容 |
|---|---|
| `dtx_linux_amd64.tar.gz` | Linux (x86_64) バイナリ |
| `dtx_linux_arm64.tar.gz` | Linux (ARM64) バイナリ |
| `dtx_darwin_amd64.tar.gz` | macOS (Intel) バイナリ |
| `dtx_darwin_arm64.tar.gz` | macOS (Apple Silicon) バイナリ |
| `dtx_windows_amd64.zip` | Windows (x86_64) バイナリ |
| `checksums.txt` | 各アーカイブの SHA-256 チェックサム |

### バージョン番号のルール

[Semantic Versioning](https://semver.org/) に従う。

* `v1.2.3` — 正式リリース（GitHub Release として公開）
* `v1.2.3-beta.1` などサフィックスあり → pre-release として自動分類される

### バージョン文字列の確認

リリースバイナリには `ldflags` でバージョンが埋め込まれる。

```bash
dtx --version  # → dtx v0.1.0
```

開発ビルド（`go build` のまま）では `dtx dev` と表示される。

### ローカルでのリリースビルド確認

GoReleaser をインストール済みの場合、ローカルで成果物を確認できる（GitHub へは publish しない）。

```bash
goreleaser release --snapshot --clean
# dist/ ディレクトリに各プラットフォームのバイナリが生成される
```

## 注意事項

* `dtx edit` は `$VISUAL`、次に `$EDITOR` を使う
* どちらも未設定の場合、`dtx edit` はエラーにする
* `dtx run` のdtx側オプションは `--` より前だけで解釈する
* `--` 以降は実行対象コマンドへそのまま渡す
* `dtx doctor`、シェル統合、鍵ローテーションはMVP対象外
