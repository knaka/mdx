package mdx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

func TestCodeBlock(t *testing.T) {
	input := bytes.NewBufferString(`Code block:

<!-- mdxcode src=misc/hello.c -->


			hello
	
			world

* foo

  <!-- mdxcode src=misc/hello.c -->

      foo

      bar

Done.
`)
	expected := []byte(`Code block:

<!-- mdxcode src=misc/hello.c -->


			#include <stdio.h>
			
			int main (int argc, char** argv) {
				printf("Hello!\n");
			}

* foo

  <!-- mdxcode src=misc/hello.c -->

      #include <stdio.h>
      
      int main (int argc, char** argv) {
      	printf("Hello!\n");
      }

Done.
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err != nil {
		t.Fatal("error")
	}
	if bytes.Compare(expected, output.Bytes()) != 0 {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(string(expected), output.String()))
	}
}

func TestFencedCodeBlock(t *testing.T) {
	input := bytes.NewBufferString(`Code block:

<!-- mdxcode src=misc/hello.c -->

~~~
hello

world
~~~

* foo

  <!-- mdxcode src=misc/hello.c -->

  ~~~
  hello
  
  world
  ~~~

Done.
`)
	expected := []byte(`Code block:

<!-- mdxcode src=misc/hello.c -->

~~~
#include <stdio.h>

int main (int argc, char** argv) {
	printf("Hello!\n");
}
~~~

* foo

  <!-- mdxcode src=misc/hello.c -->

  ~~~
  #include <stdio.h>
  
  int main (int argc, char** argv) {
  	printf("Hello!\n");
  }
  ~~~

Done.
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err != nil {
		t.Fatal("error")
	}
	if bytes.Compare(expected, output.Bytes()) != 0 {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(string(expected), output.String()))
	}
}

func TestFencedCodeBlockNotClosing(t *testing.T) {
	input := bytes.NewBufferString(`Code block:

<!-- mdxcode src=misc/hello.c -->

Done
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err == nil || err.Error() != "stack not empty" {
		t.Fatal("error")
	}
}

func TestToc(t *testing.T) {
	input := bytes.NewBufferString(`TOC:

<!-- mdxtoc pattern=misc/*.md -->
<!-- /mdxtoc -->

* foo

  <!-- mdxtoc pattern=misc/*.md -->
  foo  
  <!-- /mdxtoc -->

`)
	expected := []byte(`TOC:

<!-- mdxtoc pattern=misc/*.md -->
* [Bar ドキュメント](misc/bar.md)
* [misc/foo.md](misc/foo.md)
<!-- /mdxtoc -->

* foo

  <!-- mdxtoc pattern=misc/*.md -->
  * [Bar ドキュメント](misc/bar.md)
  * [misc/foo.md](misc/foo.md)
  <!-- /mdxtoc -->

`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err != nil {
		t.Fatal(err.Error())
	}
	if bytes.Compare(expected, output.Bytes()) != 0 {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(string(expected), output.String()))
	}
}

func TestTocDifferentDepth(t *testing.T) {
	input := bytes.NewBufferString(`TOC:

<!-- mdxtoc pattern=misc/*.md -->
* foo
* bar

other document

* foo

  <!-- /mdxtoc -->
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err == nil {
		t.Fatal("Do not succeed")
	} else {
		if !strings.HasPrefix(err.Error(), "commands do not match") {
			t.Fatal("not expected error")
		}
	}
}

func TestLinks(t *testing.T) {
	input := bytes.NewBufferString(`Links:

Inline-links <!-- mdxlink href=misc/foo.md -->...<!-- /mdxlink -->
and <!-- mdxlink href=misc/bar.md -->...<!-- /mdxlink --> works.
`)
	expected := []byte(`Links:

Inline-links <!-- mdxlink href=misc/foo.md -->[misc/foo.md](misc/foo.md)<!-- /mdxlink -->
and <!-- mdxlink href=misc/bar.md -->[Bar ドキュメント](misc/bar.md)<!-- /mdxlink --> works.
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err != nil {
		t.Fatal(err.Error())
	}
	if bytes.Compare(expected, output.Bytes()) != 0 {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(string(expected), output.String()))
	}
}

func _TestIncludes(t *testing.T) {
	input := bytes.NewBufferString(`Includes:

<!-- mdxinclude src=misc/foo.md -->
<!-- /mdxinclude -->
`)
	expected := []byte(`Includes:

<!-- mdxinclude src=misc/foo.md -->
<!-- /mdxinclude -->
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err != nil {
		t.Fatal(err.Error())
	}
	if bytes.Compare(expected, output.Bytes()) != 0 {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(string(expected), output.String()))
	}
}

func TestTitle(t *testing.T) {
	input1 := bytes.NewBufferString(`---
title: My Title
---
`)
	title := getMarkdownTitleSub(input1, "default")
	if title != "My Title" {
		t.Fatal("Could not find title")
	}
}

func TestTitle2(t *testing.T) {
	input1 := bytes.NewBufferString(`---

---
title: Foo Bar
`)
	title := getMarkdownTitleSub(input1, "default")
	if title != "default" {
		t.Fatal("How did you get it?")
	}
}

func TestTitle3(t *testing.T) {
	input1 := bytes.NewBufferString(`% My Document

Document.
`)
	title := getMarkdownTitleSub(input1, "default")
	if title != "My Document" {
		t.Fatal("Could not get title")
	}
}

func TestTitle4(t *testing.T) {
	input1 := bytes.NewBufferString(`% My document title 
 is long
Document.
`)
	title := getMarkdownTitleSub(input1, "default")
	if title != "My document title is long" {
		t.Fatal("Could not get title")
	}
}

func TestTitle5(t *testing.T) {
	input1 := bytes.NewBufferString(`Title:   My title
Author:  Foo Bar

Main document.
`)
	title := getMarkdownTitleSub(input1, "default")
	if title != "My title" {
		t.Fatal("Could not get title")
	}
}

// Unknown commands are ignored
func TestUnknown(t *testing.T) {
	input := bytes.NewBufferString(`Includes:

<!-- mdxunknown src=misc/foo.md -->
<!-- /mdxunknown -->

`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err == nil {
		t.Fatal(err.Error())
	}
}

func TestTocFail(t *testing.T) {
	input := bytes.NewBufferString(`TOC:

<!-- mdxtoc pattern=misc/*.md -->
<!-- /mdxtoc -->
<!-- /mdxtoc -->

`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err == nil {
		t.Fatal("Error")
	}
}

func TestCodeBlockWithBlankLines(t *testing.T) {
	// fails to figure out correct indent
	// library does not have meta-info of the block
	t.Skip()
	input := bytes.NewBufferString(`Code block:

* foo

  <!-- mdxcode src=misc/code_with_blank_lines.c -->

  ~~~
    
  
  #include <stdio.h>
  
  int main (int argc, char** argv) {
  printf("Hello!\n");
  }
  ~~~

<!-- /mdxcode -->
`)
	expected := []byte(`Code block:

* foo

  <!-- mdxcode src=misc/code_with_blank_lines.c -->

  ~~~
  
  
  #include <stdio.h>
  
  int main (int argc, char** argv) {
  	printf("Hello!\n");
  }
  ~~~

<!-- /mdxcode -->
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err != nil {
		t.Fatal("error")
	}
	if bytes.Compare(expected, output.Bytes()) != 0 {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(string(expected), output.String()))
	}
}
