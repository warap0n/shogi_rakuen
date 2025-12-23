# Shogi Rakuen (将棋楽園)

将棋ゲームのバックエンドAPIサーバー。将棋のルールエンジンとゲーム管理機能を提供します。

## 技術スタック

- **Language**: Go 1.23.2
- **Web Framework**: Echo v4
- **ORM**: GORM
- **Database**: SQLite (開発), PostgreSQL (本番想定)
- **Authentication**: JWT (Cookie-based)
- **Testing**: testify, mock
- **Architecture**: Clean Architecture (レイヤー分離)

## 主な機能

### 将棋ルールエンジン
- 全駒種の動き判定（歩、香、桂、銀、金、角、飛、王）
- 成り駒の処理（と金、成香、成桂、成銀、馬、龍）
- 駒の取得と持ち駒管理
- 盤面状態の管理
- 移動履歴の記録

### ゲーム管理API
- 対局の開始・取得
- 駒の移動処理
- 手順履歴の管理

### ユーザー認証
- サインアップ/ログイン
- JWT認証（Cookie）
- パスワードハッシュ化（bcrypt）

## アーキテクチャ

クリーンアーキテクチャに基づいたレイヤー構成：

```
┌─────────────────────────────────────┐
│         Controller (HTTP)           │  ← リクエスト/レスポンス処理
├─────────────────────────────────────┤
│           Usecase                   │  ← ビジネスロジック調整
├─────────────────────────────────────┤
│          Repository                 │  ← データ永続化
├─────────────────────────────────────┤
│       Model (Domain)                │  ← 将棋ルールエンジン
└─────────────────────────────────────┘
```

### ディレクトリ構成

```
.
├── controller/       # HTTPハンドラー、DTOの定義
├── usecase/          # ユースケース層、入出力DTO
├── repository/       # データアクセス層
├── model/            # ドメインモデル（将棋ルール）
├── router/           # ルーティング設定
└── main.go           # エントリーポイント
```

## セットアップ

### 前提条件

- Go 1.23.2以上
- Docker & Docker Compose (PostgreSQL使用時)

### 環境変数

`.env` ファイルを作成：

```bash
JWT_SECRET=your-secret-key
API_DOMAIN=localhost
FE_URL=http://localhost:3000
```

### ローカル実行（SQLite）

```bash
# 依存関係のインストール
go mod download

# 実行
go run main.go
```

サーバーは `http://localhost:8080` で起動します。

### Docker Compose実行（PostgreSQL）

```bash
# コンテナ起動
docker-compose up -d

# ログ確認
docker-compose logs -f shogi-api
```

## API エンドポイント

### 認証（認証不要）

| Method | Path | 説明 |
|--------|------|------|
| POST | `/signup` | 新規ユーザー登録 |
| POST | `/login` | ログイン（JWT発行） |

### ユーザー（認証必須）

| Method | Path | 説明 |
|--------|------|------|
| GET | `/me` | 自分のユーザー情報取得 |

### ゲーム（認証不要）※今後認証必須化予定

| Method | Path | 説明 |
|--------|------|------|
| POST | `/game/start` | 新規対局開始 |
| GET | `/game/:id` | 対局情報取得 |
| POST | `/game/:id/move` | 駒を動かす |
| GET | `/game/:id/moves` | 手順履歴取得 |

### リクエスト例

#### 対局開始

```bash
curl -X POST http://localhost:8080/game/start \
  -H "Content-Type: application/json" \
  -d '{
    "black_id": "player1",
    "white_id": "player2"
  }'
```

#### 駒の移動

```bash
curl -X POST http://localhost:8080/game/:id/move \
  -H "Content-Type: application/json" \
  -d '{
    "from": {"rank": 6, "file": 6},
    "to": {"rank": 5, "file": 6},
    "promote": false,
    "drop": false
  }'
```

座標系：
- `rank`: 段（0-8、0が1段目）
- `file`: 筋（0-8、0が1筋目）

## テスト

### 全テスト実行

```bash
go test ./... -v
```

### カバレッジレポート生成

```bash
./cover.sh
```

カバレッジレポートが `coverage.html` に出力されます。

### テスト構成

- **ユニットテスト**: model, usecase層のロジックテスト
- **統合テスト**: repository層のDB操作テスト
- **コントローラーテスト**: HTTPハンドラーのテスト（モック使用）

## 開発状況

### 実装済み
- ✅ 将棋の基本ルール（駒の動き、成り、持ち駒）
- ✅ ゲーム管理API（開始、移動、履歴）
- ✅ ユーザー認証（JWT）
- ✅ レイヤー分離アーキテクチャ
- ✅ 単体テスト/統合テスト

### 未実装（今後の拡張）
- ⬜ 二歩・王手・詰みの判定
- ⬜ リアルタイム対局（WebSocket）
- ⬜ レーティングシステム
- ⬜ 対局履歴の保存・再生
- ⬜ AI対戦機能

## ライセンス

MIT
