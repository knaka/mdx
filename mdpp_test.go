package mdpp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

func TestCodeBlock(t *testing.T) {
	input := bytes.NewBufferString(`Code block:

<!-- mdppcode src=misc/hello.c -->


			hello
	
			world

* foo

  <!-- mdppcode src=misc/hello.c -->

      foo

      bar

Done.
`)
	expected := []byte(`Code block:

<!-- mdppcode src=misc/hello.c -->


			#include <stdio.h>
			
			int main (int argc, char** argv) {
				printf("Hello!\n");
			}

* foo

  <!-- mdppcode src=misc/hello.c -->

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

<!-- mdppcode src=misc/hello.c -->

~~~
hello

world
~~~

* foo

  <!-- mdppcode src=misc/hello.c -->

  ~~~
  hello
  
  world
  ~~~

Done.
`)
	expected := []byte(`Code block:

<!-- mdppcode src=misc/hello.c -->

~~~
#include <stdio.h>

int main (int argc, char** argv) {
	printf("Hello!\n");
}
~~~

* foo

  <!-- mdppcode src=misc/hello.c -->

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

<!-- mdppcode src=misc/hello.c -->

Done
`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err == nil || err.Error() != "stack not empty" {
		t.Fatal("error")
	}
}

func TestToc(t *testing.T) {
	input := bytes.NewBufferString(`TOC:

<!-- mdpptoc pattern=misc/*.md -->
<!-- /mdpptoc -->

* foo

  <!-- mdpptoc pattern=misc/*.md -->
  foo  
  <!-- /mdpptoc -->

`)
	expected := []byte(`TOC:

<!-- mdpptoc pattern=misc/*.md -->
* [Bar ドキュメント](misc/bar.md)
* [misc/foo.md](misc/foo.md)
<!-- /mdpptoc -->

* foo

  <!-- mdpptoc pattern=misc/*.md -->
  * [Bar ドキュメント](misc/bar.md)
  * [misc/foo.md](misc/foo.md)
  <!-- /mdpptoc -->

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

<!-- mdpptoc pattern=misc/*.md -->
* foo
* bar

other document

* foo

  <!-- /mdpptoc -->
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

Inline-links <!-- mdpplink href=misc/foo.md -->...<!-- /mdpplink -->
and <!-- mdpplink href=misc/bar.md -->...<!-- /mdpplink --> works.
`)
	expected := []byte(`Links:

Inline-links <!-- mdpplink href=misc/foo.md -->[misc/foo.md](misc/foo.md)<!-- /mdpplink -->
and <!-- mdpplink href=misc/bar.md -->[Bar ドキュメント](misc/bar.md)<!-- /mdpplink --> works.
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

<!-- mdppinclude src=misc/foo.md -->
<!-- /mdppinclude -->
`)
	expected := []byte(`Includes:

<!-- mdppinclude src=misc/foo.md -->
<!-- /mdppinclude -->
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

<!-- mdppunknown src=misc/foo.md -->
<!-- /mdppunknown -->

`)
	output := bytes.NewBuffer(nil)
	if err := PreprocessWithoutDir(output, input); err == nil {
		t.Fatal(err.Error())
	}
}

func TestTocFail(t *testing.T) {
	input := bytes.NewBufferString(`TOC:

<!-- mdpptoc pattern=misc/*.md -->
<!-- /mdpptoc -->
<!-- /mdpptoc -->

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

  <!-- mdppcode src=misc/code_with_blank_lines.c -->

  ~~~
    
  
  #include <stdio.h>
  
  int main (int argc, char** argv) {
  printf("Hello!\n");
  }
  ~~~

<!-- /mdppcode -->
`)
	expected := []byte(`Code block:

* foo

  <!-- mdppcode src=misc/code_with_blank_lines.c -->

  ~~~
  
  
  #include <stdio.h>
  
  int main (int argc, char** argv) {
  	printf("Hello!\n");
  }
  ~~~

<!-- /mdppcode -->
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
