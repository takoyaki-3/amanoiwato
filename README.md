# 天岩戸

天岩戸（あまのいわと）は、作業中ファイルやシステム用ファイル以外の半永久的に保存するファイルを管理するストレージである。

- projects: コンテスト応募などのプロジェクトにおいて半永久的に保存するファイル
- texts: 動機や反省、文章化した時点での思考整理結果に加え、受験や就職活動で自らの考えを整理し文章化したファイル
- reports: 卒論やプロジェクトの結果が分かる１ファイル

## 使い方

#### 前提条件
Golangがインストールされた環境が前提となる。

ストレージ内のファイルに対し同期を取るには、次の手順に従う。

1. 環境変数の設定

S3互換ストレージにおける環境変数を`.env`ファイルとして設定する。

```env
S3_BUCKET=<your_bucket_name>
AWS_ACCESS_KEY_ID=<your_access_key>
AWS_SECRET_ACCESS_KEY=<your_secret_access_key>
AWS_REGION=<your_aws_region>
S3_ENDPOINT=<your_s3_endpoint>
```

2. コマンドの実行

```sh
go run sync.go
```

以上でストレージ内のファイルが同期される。

