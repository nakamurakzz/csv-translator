# csv-translator

このツールは、CSV ファイルの内容を Google Cloud Translation API を使用して英語に翻訳するためのコマンドラインツールです。

## 機能

- CSV ファイルの各セルを英語に翻訳
- 特定のカラムを翻訳対象から除外可能
- 翻訳結果のキャッシュ機能
- 翻訳結果を新しい CSV ファイルとして出力

## 必要条件

- Go 1.24 以上
- Google Cloud Platform のアカウント
- Google Cloud Translation API の有効化
- 環境変数 `GOOGLE_CLOUD_PROJECT` の設定

## インストール

```bash
go get github.com/yourusername/csv-translator
```

## 使用方法

```bash
./csv-translator <input.csv> <exclude_cols>
```

### パラメータ

- `input.csv`: 翻訳対象の CSV ファイル
- `exclude_cols`: 翻訳対象から除外するカラム名（カンマ区切り）

### 例

```bash
./csv-translator data.csv "id,tel,postal"
```

この例では、`data.csv`ファイルを翻訳し、`id`、`tel`、`postal`カラムは翻訳対象から除外されます。

## 出力

翻訳された CSV ファイルは、元のファイル名に`_translated`を付加した名前で保存されます。
例：`data.csv` → `data_translated.csv`

## 注意事項

- 空のセルは翻訳されません
- 翻訳エラーが発生した場合、元のテキストがそのまま保持されます
- 同じテキストの重複翻訳を避けるため、キャッシュ機能が実装されています

## ライセンス

MIT License
