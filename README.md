# slack-chatgpt-bot

このリポジトリは2023年 CyberAgent Developers Advent Calendar 2023 5日目の記事で紹介した、Slack上で動作するチャットボットのソースコードです。

[ゲーム開発でどう活かす？ChatGPTをみんなで使える仕組みづくり](https://developers.cyberagent.co.jp/blog/archives/44803/)

メンテナンスの予定はなく、不具合や機能追加のリクエスト、プルリクエストには対応しません。

## ライセンス

CC0 1.0 Universal (CC0 1.0) Public Domain Dedication

また、このソースコードを利用したことによるいかなる損害についても、作者・CyberAgentは一切の責任を負いません。

## 使い方

**Slackアプリケーションの作成**

[https://api.slack.com/apps](https://api.slack.com/apps) から新たにアプリケーションを作成し、以下の権限を付与してください。

```
# App-Level Tokens
connections:write

# Bot Token Scopes
app_mentions:read

# パブリックチャンネルで動作させる場合は以下が必要
channels:history
channels:read
chat:write
# プライベートチャンネルで動作させる場合は以下が必要
groups:history 
groups:read
# ダイレクトメッセージで動作させる場合は以下が必要
im:history
im:read
im:write
```

また、「Socket Mode」を有効にしてください。

**OpenAI API Keyの取得**

https://platform.openai.com/api-keys からAPI Keyを取得してください。

**設定の記入**

環境変数または `.env` ファイルによって設定します。

```
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxxxxxxx
SLACK_APP_LEVEL_TOKEN=xapp-1-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_ORGANIZATION_ID=org-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx # 任意
OPENAI_MODEL=gpt-4
LOG_LEVEL=INFO

# スプレッドシートによる統計情報の記録を行う場合は以下を設定
GOOGLE_APPLICATION_CREDENTIALS_JSON=
GOOGLE_SERVICE_ACCOUNT_EMAIL=
SPREADSHEET_ID=
```

**システムメッセージの編集**
システムメッセージを指定するファイルは `config/system.txt` です。ビルド時点のものが利用されます。

記事中では紹介しませんでしたが、 `{{custom_instructions}}` の部分をチャンネルの説明・トピックで指定した文章で置き換える機能があります。 デフォルトで `SlackBot:` 以降の改行までの文章が自動的にロードされます。詳しくは `slackapi/api.go` を参照してください。

**起動**

Dockerfileをビルドして起動するか、Go言語の実行環境を用意して起動してください。

```
$ docker build -t slack-chatgpt-bot .
$ docker run -it --rm --env-file .env slack-chatgpt-bot
```

```
$ go run main.go wire_gen.go
```

HTTP等のポートは動作には必要ありませんが、死活監視用に `:8080` もしくはPORT環境変数で指定したポートで常に `200 OK` を返すようになっています。

ログメッセージに Hello Slack! と表示されれば起動完了です。作成したBotユーザーに対しメンションをすることでBotが動作します。
