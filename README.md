---
title: Document for mdx(1)
---

mdx(1)

> Japanese version is here: <!-- mdxlink href=./README-ja.md -->[mdx(1) ドキュメント（日本語）](./README-ja.md)<!-- /mdxlink -->

# NAME

mdx - Markdown preprocessor for resolving cross-references between files

# INSTALLATION

    $ go install github.com/knaka/mdx/cmd/mdx@latest

# SYNOPSIS

Concatenate the rewritten results and output to standard output.

    mdx input1.md input2.md > output.md

In-place rewriting.

    mdx -i rewritten1.md rewritten2.md

# DESCRIPTION

If code from another file is inserted into a code block in a Markdown document and then the original code is rewritten, the inserted code does not automatically reflect the rewritten code. Also, if you create index of Markdown documents, any increase or decrease of the document will not be reflected. In general, it is not possible to resolve cross-references between files.

The command mdx(1) is assumed to be called from a Makefile or similar. It rewrites the input according to the metacommands in the comments included in the input, and outputs it to the output file.

The command mdx(1) with the `-i` (`--in-place`) option will rewrite the files in-place. It is intended to be set in the editor's “Program to be executed on save” or similar.

When the code in the code block have to be rewritten to the latest content, the follwing input will give:

    <!-- mdxcode src=src/hello.c -->

        foo
        bar

the following output. The metacommands in the output remain as they were in the input, so the output can be input again. Indented code blocks and fenced code blocks works.

    <!-- mdxcode src=src/hello.c -->

        #include <stdio.h>

        int main(int argc, char** argv) {
            printf("Hello, World!");
            return (0);
        }

When mdx(1) updates the Markdown listing of the files in a directory, the following input will:

    <!-- mdxtoc pattern=docs/*.md -->
    * [Already deleted document](docs/deleted.md)
    * [Hello document](docs/hello.md)
    * [World document](docs/world.md)
    <!-- /mdxtoc -->

make the following output. Supported style for writing titles are YAML metadata, Pandoc title blocks, and MultiMarkdown style. If the file itself is included in the list, it will not be a link.

    <!-- mdxtoc pattern=docs/*.md -->
    * [Hello document](docs/hello.md)
    * [New document](docs/new.md)
    * [World document](docs/world.md)
    <!-- /mdxtoc -->

As an example of an in-place setting, VSCode's plugin “[Run on Save](https://marketplace.visualstudio.com/items?itemName=pucelle.run-on-save)” will automatically run when saving a Markdown file. To run it automatically when saving a Markdown file, the following settings are used.

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

Link to markdown. Input:

    It is described in the “<!-- mdxlink href=hello.md -->...<!-- /mdxlink -->.”

Output:

    It is described in the “<!-- mdxlink href=hello.md -->Hello Document<!-- /mdxlink -->.”

# ToDo

* [ ] Inclusion command (includes header, footer...)
* [ ] Code block for files which starts with blank or indented lines
