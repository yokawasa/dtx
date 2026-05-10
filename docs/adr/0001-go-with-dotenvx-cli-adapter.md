# ADR 0001: Go + dotenvx CLI adapterで実装する

## Status

Accepted

## Context

dtxは、ローカル環境における環境変数を安全に管理し、選択したenvを `dtx run` 経由で明示的に利用するためのCLIツールである。

設計上、dtxの主な責務は以下である。

* envの選択状態を管理する
* 暗号化されたenvファイルと鍵を `~/.dtx` 配下で管理する
* 誤った環境での実行を防ぐため、実行経路を `dtx run` に限定する
* 暗号化・復号の仕様はdotenvxに委ねる

当初は、dotenvxのライブラリAPIを利用する前提でNode.js + TypeScript実装を検討していた。しかし、CLIツールとしての配布性、単一バイナリ化、プロセス実行やファイル権限の扱いやすさを考えると、dtx本体はGoで実装する方が適している。

一方で、dotenvxはNode.jsエコシステムのツールであり、GoからdotenvxのライブラリAPIを直接利用するのは自然ではない。また、dotenvx互換の暗号化・復号処理をGoで再実装すると、dtxが本来持つべきでない暗号仕様への追従責務を負うことになる。

## Decision

dtx本体はGoで実装する。

暗号化・復号・実行時注入はGoで再実装せず、dotenvx CLIをサブプロセスとして利用する。

dotenvx CLI呼び出しは、dtx本体の各コマンドに直接散らさず、adapter層に閉じ込める。

具体的には以下を採用する。

* Go moduleとして実装する
* CLIエントリポイントは `cmd/dtx/main.go` とする
* dotenvx CLIは必須外部依存とする
* `internal/dotenvx` にadapterを置く
* adapterが `run` / `encrypt` / `decrypt` 相当のdotenvx CLI呼び出しを担当する
* dtx本体はdotenvxの暗号形式を解釈しない
* dtx本体はenv選択、path解決、鍵管理、権限設定、出力制御、エラー整形を担当する

## Consequences

### Positive

* dtxを単一バイナリとして配布しやすい
* ユーザー環境にNode.jsランタイムを要求しない
* Goの標準機能でファイル権限、プロセス実行、終了コード伝播を扱いやすい
* dtxの責務をenv管理と実行ゲートに集中できる
* dotenvxの暗号仕様をdtx側で再実装せずに済む
* 将来、dotenvx CLI以外の実装へ差し替える場合もadapter層を境界にできる

### Negative

* dotenvx CLIが実行環境にインストールされている必要がある
* dotenvx CLIの仕様変更や出力変更の影響を受ける
* サブプロセス呼び出しのため、ライブラリAPI利用よりも制御できる範囲が狭い
* dotenvx由来の標準出力と実行対象コマンドの標準出力を分離する設計に注意が必要

### Neutral / Mitigation

* dotenvx CLIの存在確認を、dotenvx利用コマンドの実行前に行う
* dotenvx CLI呼び出しはadapter層に閉じ込め、仕様変更の影響範囲を限定する
* `--verbose` 指定時のみdotenvx由来の詳細出力を表示する
* 通常時は `--quiet` 相当の出力抑制を行う
* dtxのエラーメッセージはdotenvxの生エラーをそのまま出さず、dtxの文脈で整形する

## Alternatives Considered

### Node.js + TypeScript + dotenvx library API

dotenvxとの統合は自然だが、dtx本体をNode.jsランタイム前提にする必要がある。

CLI配布、単一バイナリ化、ファイル権限やプロセス制御の扱いやすさではGo実装の方が適しているため採用しない。

### Goでdotenvx互換の暗号処理を再実装する

外部CLI依存をなくせるが、dotenvxの暗号形式や鍵管理仕様へ追従する必要がある。

dtxの本質は暗号ライブラリではなくenv管理と実行ゲートであるため、MVPでは採用しない。

### Go + dotenvx CLIを各コマンドから直接呼び出す

実装は短くなるが、dotenvx CLI依存がコード全体に広がる。

将来の差し替えやテストが難しくなるため、adapter層に閉じ込める方針を採用する。

## Follow-ups

* dotenvx CLIへ `~/.dtx/keys/<env>` の秘密鍵を渡す具体方式を実装時に確定する
* `dotenvx run` のquiet/verbose挙動を実測し、dtxの出力制御仕様に反映する
* adapter層のテストでは、実dotenvx CLIを使うテストとfake adapterを使うテストを分ける
