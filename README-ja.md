---
title: mdx(1) ドキュメント（日本語）
---

mdx(1)

# NAME

mdx - ファイル間の相互参照解決のための Markdown プリプロセッサ

# INSTALLATION

    $ go install github.com/knaka/mdx/cmd/mdx@latest

# SYNOPSIS

書き換えた結果を連結して標準出力へ出力

    mdx input1.md input2.md > output.md

インプレース書き換え

    mdx -i rewritten1.md rewritten2.md

# DESCRIPTION

コマンド mdx(1) は Makefile などから呼ばれることを想定しています。入力に含まれるコメントに従って入力を書き換え、出力ファイルへ出力します。

コマンド mdx(1) に `-i` (`--in-place`) オプションをつけた場合は、インプレースでファイルを書き換えます。エディタの「セーブ時に実行されるスクリプト」などに設定されることを想定しています。

コードブロック内コードを最新の内容に書き換え:

入力

    <!-- mdxcode src=src/hello.c -->

        foo
        bar

出力

    <!-- mdxcode src=src/hello.c -->

        #include <stdio.h>

        int main(int argc, char** argv) {
            printf("Hello, World!");
            return (0);
        }

ディレクトリ内の Markdown の一覧を更新:

    <!-- mdxtoc pattern=docs/*.md -->
    * [Already deleted document](docs/deleted.md)
    * [Hello document](docs/hello.md)
    * [World document](docs/world.md)
    <!-- /mdxtoc -->

出力

    <!-- mdxtoc pattern=docs/*.md -->
    * [Hello document](docs/hello.md)
    * [New document](docs/new.md)
    * [World document](docs/world.md)
    <!-- /mdxtoc -->

VSCode の [Run on Save](https://marketplace.visualstudio.com/items?itemName=pucelle.run-on-save) プラグインで、Markdown ファイルをセーブする際に自動的に更新する設定の例。

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

# OPTIONS

