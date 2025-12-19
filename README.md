# csvfilter

CSVファイルを読み込み、部分一致フィルタと指定列ソートを行うCLIツールです。

## 使い方

```bash
# サンプルCSVを使う
csvfilter -file sample.csv -filter 名前 -value 山田

# 引数1つ目をファイルとして扱う
csvfilter sample.csv -sort 年齢 -type ge

# 標準入力
cat sample.csv | csvfilter -filter 名前 -value 山田 -sort 年齢 -type lt

# 数値/文字列を明示指定
csvfilter sample.csv -sort 年齢 -type asc:num
csvfilter sample.csv -sort 名前 -type desc:str
```

## オプション

- `-file` : CSVファイルパス。省略時は引数1つ目、さらに無ければ標準入力。
- `-filter` : フィルタ対象の列名（ヘッダ名）。
- `-value` : 部分一致に使う文字列。
- `-sort` : ソート対象の列名（ヘッダ名）。
- `-type` : ソート方向。`asc/desc` または `lt/le/gt/ge`。`:`で比較方法を指定可能（例: `asc:num`, `desc:str`）。

## 仕様

- 1行目はヘッダとして扱います。
- フィルタは部分一致（`strings.Contains`）です。
- ソートは数値として解釈できる場合は数値比較、それ以外は文字列比較です。
- `:num` を指定した場合、対象列は全て数値である必要があります。
- 出力はCSV形式で標準出力に書き出します。

## ビルド

```bash
go build -o csvfilter .
```

## サンプル

`sample.csv` を同梱しています。

## ライセンス

MIT License. See `LICENSE`.
