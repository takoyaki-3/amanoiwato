# 天岩戸

天岩戸（あまのいわと）は、作業中ファイルやシステム用ファイル以外の半永久的に保存するファイルを管理するためのストレージ同期ツールです。

## 概要

天岩戸は、指定されたローカルディレクトリとS3互換ストレージ（例：Amazon S3、MinIOなど）間でファイルの同期を行います。 
これにより、重要なファイルのバックアップや複数デバイス間での共有を容易に行うことができます。

- コンテスト応募などのプロジェクトファイル（projects）
- 動機や反省、思考整理結果、受験や就職活動で文章化したファイル（texts）
- 卒論やプロジェクトの結果が分かるファイル（reports）

などを安全に保管・管理できます。

## 特徴

- S3互換ストレージとの双方向同期
- Dockerによる容易な環境構築
- Pythonによるシンプルな実装

## ファイルツリー

```
.
├── Dockerfile
├── README.md
├── docker-compose.yml
├── requirements.txt
└── sync.py
```

- **Dockerfile:** Dockerイメージ構築のための設定ファイル
- **README.md:** このドキュメント
- **docker-compose.yml:** Dockerコンテナの設定ファイル
- **requirements.txt:** Pythonパッケージの依存関係リスト
- **sync.py:** S3との同期処理を行うPythonスクリプト

## インストール方法

1. DockerおよびDocker Composeがインストールされていることを確認してください。
2. このリポジトリをクローンします。
3. `.env`ファイルを作成し、以下の環境変数を設定します。

```env
S3_BUCKET=<your_bucket_name>
AWS_ACCESS_KEY_ID=<your_access_key>
AWS_SECRET_ACCESS_KEY=<your_secret_access_key>
AWS_REGION=<your_aws_region>
S3_ENDPOINT=<your_s3_endpoint>
LOCAL_DIRECTORY=<your_local_directory>
BUCKET_NAME=<your_bucket_name> 
```

- `S3_BUCKET`: S3バケット名
- `AWS_ACCESS_KEY_ID`: AWSアクセスキーID
- `AWS_SECRET_ACCESS_KEY`: AWSシークレットアクセスキー
- `AWS_REGION`: AWSリージョン
- `S3_ENDPOINT`: S3エンドポイント（S3互換ストレージを使用する場合は必須）
- `LOCAL_DIRECTORY`: ローカルで同期するディレクトリ
- `BUCKET_NAME`: S3バケット名


## 使い方

1. Dockerコンテナを起動します。

```sh
docker-compose up -d
```

2.  コンテナ内で同期スクリプトを実行します。

```sh
docker-compose exec sync_service python sync.py
```

これにより、`LOCAL_DIRECTORY`で指定したローカルディレクトリとS3バケット間でファイルの同期が実行されます。

## APIエンドポイント

天岩戸はAPIエンドポイントを提供していません。 

## 設定ファイル

- `.env`: 環境変数の設定ファイル
- `docker-compose.yml`: Dockerコンテナの設定ファイル

## コマンド実行例

- Dockerコンテナのビルドと起動: `docker-compose up -d`
- 同期スクリプトの実行: `docker-compose exec sync_service python sync.py`

## ライセンス

このプロジェクトはMITライセンスで公開されています。