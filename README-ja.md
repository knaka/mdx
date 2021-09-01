---
title: mdx(1) ドキュメント（日本語）
---

mdx(1)

# NAME

mdx - ファイル間の相互参照を解決するための Markdown プリプロセッサ

# INSTALLATION

    $ go install github.com/knaka/mdx/cmd/mdx@latest

# SYNOPSIS

書き換えた結果を連結して標準出力へ出力する。

    mdx input1.md input2.md > output.md

インプレースで書き換える。

    mdx -i rewritten1.md rewritten2.md

# DESCRIPTION

Markdown 文書のコードブロックに他のファイルのコードを取り込んだ後で元のコードが書き換えられても、取り込んだコードへ自動で反映させることはできません。また、Markdown 文書の目録を記載する際にも、文書が増減してもそれが反映されません。総じて、ファイル間の相互参照を解決することができないのです。

コマンド mdx(1) は Makefile などから呼ばれることを想定しています。入力に含まれるコメント内のメタコマンドに従って入力を書き換え、出力ファイルへ出力します。

コマンド mdx(1) に `-i` (`--in-place`) オプションをつけた場合は、インプレースでファイルを書き換えます。エディタの「セーブ時に実行されるスクリプト」などに設定されることを想定しています。

コードブロック内コードを最新の内容に書き換える場合には、下記のような入力に対して:

    <!-- mdxcode src=src/hello.c -->

        foo
        bar

以下のような出力をする。出力内のメタコマンドは入力の通りに残っているので、これは再度入力にすることができる。インデント型のコードブロック、フェンス型のコードブロックともに有効である。

    <!-- mdxcode src=src/hello.c -->

        #include <stdio.h>

        int main(int argc, char** argv) {
            printf("Hello, World!");
            return (0);
        }

ディレクトリ内の Markdown の一覧を更新する場合には、例えば下記のような入力に対し:

    <!-- mdxtoc pattern=docs/*.md -->
    * [Already deleted document](docs/deleted.md)
    * [Hello document](docs/hello.md)
    * [World document](docs/world.md)
    <!-- /mdxtoc -->

下記のような出力を行う。タイトルの記述方法しとては、YAML メタデータ、Pandoc タイトルブロック、MultiMarkdown スタイルに対応している。自身のファイルが一覧に含まれる場合は、それはリンクにならない。

    <!-- mdxtoc pattern=docs/*.md -->
    * [Hello document](docs/hello.md)
    * [New document](docs/new.md)
    * [World document](docs/world.md)
    <!-- /mdxtoc -->

In-place での設定例としては、VSCode の [Run on Save](https://marketplace.visualstudio.com/items?itemName=pucelle.run-on-save) プラグインでは、Markdown ファイルをセーブする際に自動的に実行するには、下記のような設定になる。

    "runOnSave.commands": [
        {
            "match": ".*\\.md$",
            "command": "mdx --in-place ${file}",
            "runIn": "backend",
            "runningStatusMessage": "Rewriting: ${fileBasename}",
            "finishStatusMessage": "Done: ${fileBasename}"
        },
        {}
    ],

Markdown へのリンク。入力:

    それについては「<!-- mdxlink href=hello.md -->...<!-- /mdxlink -->」に記述されている。

出力:

    それについては「<!-- mdxlink href=hello.md -->ハロー文書<!-- /mdxlink -->」に記述されている。


# OPTIONS

```
  -h, --help             Show Help
  -i, --in-place         Edit file(s) in place
  -o, --outfile string   Output outFile
```
