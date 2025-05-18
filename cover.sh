#!/usr/bin/env bash
set -euo pipefail

# 対象パッケージのリスト（必要に応じて追加）
choices=(
  "controller"
  "repository"
  "usecase"
  "model"
  "all"
  "exit"
)

echo "=== Coverage Report Generator ==="
PS3="測定したいパッケージを選んでください（番号を入力）: "

select choice in "${choices[@]}"; do
  case $choice in
    controller)
      pkgs="./controller"
      break
      ;;
    repository)
      pkgs="./repository"
      break
      ;;
    usecase)
      pkgs="./usecase"
      break
      ;;
    model)
      pkgs="./model"
      break
      ;;
    all)
      pkgs="./..."
      break
      ;;
    exit)
      echo "終了します"
      exit 0
      ;;
    *)
      echo "無効な選択です。もう一度入力してください。"
      ;;
  esac
done

echo "=== [$pkgs] のテストを実行し、coverage.out を生成します ==="
go test $pkgs -coverprofile=coverage.out

echo "=== HTML レポートを生成します ==="
go tool cover -html=coverage.out -o coverage.html

echo "=== coverage.html を開きます ==="
if command -v open >/dev/null 2>&1; then
  open coverage.html
elif command -v xdg-open >/dev/null 2>&1; then
  xdg-open coverage.html
else
  echo "生成完了: coverage.html をブラウザで開いてください"
fi
