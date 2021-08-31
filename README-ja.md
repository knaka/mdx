# mdx(1)

## 名称

mdx - ファイル間の相互参照解決のための Markdown プリプロセッサ

## 解説

コマンド mdx(1) は Makefile などから呼ばれることを想定しています。入力に含まれるコメントに従って入力を書き換え、出力ファイルへ出力します。

コマンド mdxi(1) はインプレースでファイルを書き換えます。エディタの「セーブ時に実行されるスクリプト」などに設定されることを想定しています。

コードブロック内コードを最新の内容に書き換え:

入力

    ...
    <!-- MdxReplaceCode(hello.c) -->

    ```
    foo
    bar
    ```
    ...

出力

    ...
    <!-- MdxReplaceCode(hello.c) -->

    ```
    #include <stdio.h>

    int main(int argc, char** argv) {
        printf("Hello, World!");
        return (0);
    }
    ```
    ...

ディレクトリ内の Markdown の一覧を更新:

    <!-- MdxToc(docs/*.md) { -->
    * [Already deleted content](docs/deleted.md)
    <!-- } -->

出力

    <!-- MdxToc(docs/*.md) { -->
    * [Hello document](docs/hello.md)
    * [World document](docs/world.md)
    <!-- } -->



### インストール

    $ go install github.com/knaka/mdx/cmd/mdx@latest
    $ go install github.com/knaka/mdx/cmd/mdxi@latest

## 要約

    mdx # input from stdin and output to stdout
    mdx -o <file> <files> # input from files and output to file

<!--mdxtoc pattern=*.md-->
* [README-ja.md](README-ja.md)
* [Document for MDX](README.md)
<!--/mdxtoc-->
