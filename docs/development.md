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

## 注意事項

* `dtx edit` は `$VISUAL`、次に `$EDITOR` を使う
* どちらも未設定の場合、`dtx edit` はエラーにする
* `dtx run` のdtx側オプションは `--` より前だけで解釈する
* `--` 以降は実行対象コマンドへそのまま渡す
* `dtx doctor`、シェル統合、鍵ローテーションはMVP対象外
