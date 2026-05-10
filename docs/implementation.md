# dtx 実装メモ

このドキュメントは `docs/design.md` を実装へ落とすための前提、MVPスコープ、実装方針を整理する。

## 1. 実装方針

### 1.1 実装言語

MVPは Go で実装する。

理由：

* CLIツールとして単一バイナリ配布しやすい
* ファイル権限、プロセス実行、終了コード処理を扱いやすい
* ユーザー環境にNode.jsランタイムを要求しなくてよい
* `~/.dtx` 配下のローカル状態管理と相性がよい
* テスト時に一時ディレクトリを使ったCLI検証がしやすい

注意点：

* dotenvxの暗号化・復号仕様はGoで再実装しない
* dotenvxは必須外部CLI依存とする
* dtx本体はGoで実装し、dotenvx CLI呼び出しはadapter層に閉じ込める

### 1.2 dotenvx連携

Go実装では、dotenvxのライブラリAPIではなくdotenvx CLIをサブプロセスとして利用する。

方針：

* dotenvx CLIは必須依存とする
* 起動時またはdotenvx利用コマンド実行時に `dotenvx` の存在を検証する
* dtx本体から直接dotenvxコマンドを散らさず、`internal/dotenvx` adapterに閉じ込める
* 将来、Goネイティブ実装や別の暗号バックエンドへ差し替えられる構造にする

### 1.3 Go module

MVPではGo moduleを前提にする。

想定：

```bash
go mod init github.com/yokawasa/dtx
```

CLIのエントリポイントは `cmd/dtx/main.go` とする。

## 2. MVPスコープ

MVPで実装するコマンド：

```bash
dtx use <env>
dtx current
dtx ls
dtx run [env] [--verbose] -- <command>
dtx edit <env>
```

MVPで実装しないもの：

* 鍵ローテーション
* CI向けの秘密鍵環境変数注入
* OSキーチェーン連携
* パスフレーズ方式
* プロジェクトディレクトリごとのcurrent
* direnvのような自動切り替え
* プロンプト表示の本体実装
* `dtx doctor`
* `dtx init shell`
* `dtx init completion`

## 3. 実装前提

### 3.1 `dtx edit <env>` は新規env作成も兼ねる

設計上、env作成専用コマンドは定義されていない。

MVPでは以下とする。

* `dtx edit <env>` は、対象envが存在しない場合に新規作成する
* 新規作成時はenvごとの鍵も同時に作成する
* 既存envの場合は既存鍵を再利用する

理由：

* コマンドセットを増やさずにenv作成フローを完結できる
* `edit` の直感に合う

### 3.2 `dtx doctor` はMVP対象外

`docs/design.md` では鍵権限チェックの文脈で `dtx doctor` が言及されているが、CLI仕様のコマンド一覧には含まれていない。

MVPでは以下とする。

* `dtx doctor` は実装しない
* 権限チェックは各コマンド実行時に必要最小限で行う
* `doctor` は将来の診断コマンドとして別途追加する

### 3.3 シェル統合コマンドはMVP対象外

`dtx init shell` / `dtx init completion` は将来の追加事項とする。

MVPでは以下とする。

* 補完やプロンプト表示はサンプルスクリプトに留める
* dtx本体にはシェル統合コマンドを実装しない

### 3.4 dotenvx CLI adapter

adapterは、dtx本体とdotenvx CLIの境界を担当する。

責務：

* `dotenvx` コマンドの存在確認
* envファイルパスと鍵の受け渡し
* `run` / `encrypt` / `decrypt` 相当の呼び出し
* `--quiet` / `--verbose` 相当の出力制御
* dotenvx由来のエラーをdtx向けのエラーへ変換

実装時に確認すること：

* dtxの `~/.dtx/keys/<env>` をdotenvx CLIへ渡す方法
* 一時的な `.env.keys` ファイルを作る必要があるか
* `decrypt --stdout` / `encrypt --stdout` を使えるか
* `dotenvx run` で対象コマンドの標準出力を維持したままdotenvxログだけ抑制できるか
* dotenvx CLIの終了コードとエラーメッセージの扱い

重要な制約：

* dtx本体はdotenvxの暗号形式を解釈しない
* dotenvx CLI呼び出しの詳細はadapter外へ漏らさない

### 3.5 `--verbose` の位置

MVPでは、dtxのオプションは `--` より前だけで解釈する。

```bash
dtx run --verbose -- npm start
dtx run prod --verbose -- npm start
```

`--` 以降は実行対象コマンドにそのまま渡す。

### 3.6 危険envの確認

設計ではprodなどの危険envで確認可能とされているが、具体ルールは未定義。

MVPでは以下とする。

* 確認プロンプトは実装しない
* `Using env: prod` の表示のみ行う
* 将来 `confirm` や `protected env` の設定を追加する

### 3.7 エディタ選択

MVPでは以下とする。

* `$VISUAL` を優先
* 次に `$EDITOR`
* どちらも未設定ならエラーにする

理由：

* 暗黙に `vi` などを起動すると環境によって体験がぶれる
* ユーザーに明示的なエディタ設定を促せる

### 3.8 テスト用のhome切り替え

MVPでは `DTX_HOME` をサポートする。

* 通常は `~/.dtx` を使う
* `DTX_HOME` が設定されている場合は、そのディレクトリをdtx homeとして使う

例：

```bash
DTX_HOME=/tmp/dtx-test dtx ls
```

理由：

* 実ユーザーの `~/.dtx` を汚さずテストできる
* E2Eテストが書きやすい

## 4. ファイル構成案

```text
go.mod
go.sum
cmd/
  dtx/
    main.go
internal/
  cli/
    cli.go
    parse.go
  command/
    use.go
    current.go
    ls.go
    run.go
    edit.go
  core/
    paths.go
    permissions.go
    env_name.go
    key_provider.go
    env_store.go
  dotenvx/
    adapter.go
    command.go
  process/
    runner.go
  output/
    output.go
  errors/
    errors.go
testdata/
```

## 5. 実装順序

1. Go moduleとCLIエントリポイントを作成
2. `DTX_HOME` 対応を含むpath管理を実装
3. env名バリデーションと権限設定を実装
4. `use` / `current` / `ls` を実装
5. dotenvx CLI adapterを実装
6. `edit` を実装
7. `run` を実装
8. エラー整形と出力制御を追加
9. READMEに最小利用例を追加
10. コマンド単位のテストを追加

## 6. 現時点の決定

実装を前に進めるため、以下をMVPの前提として採用する。

* Goで実装する
* Go moduleを使う
* dotenvxは必須外部CLI依存とする
* dotenvx CLI呼び出しはadapter層に閉じ込める
* `edit` は既存env編集と新規env作成を兼ねる
* `doctor` はMVPでは実装しない
* `init shell` / `init completion` はMVPでは実装しない
* `run` はGoのプロセス実行機能でdotenvx CLIを起動し、終了コードを伝播する
* `--verbose` は `--` より前でのみ解釈する
* 危険envの確認プロンプトはMVPでは実装しない
* エディタは `$VISUAL` / `$EDITOR` のみ使い、未設定ならエラーにする
* テスト用に `DTX_HOME` をサポートする
