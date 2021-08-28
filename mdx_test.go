package mdx

import (
	"regexp"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

func TestMdx(t *testing.T) {
	inputString := `A C source follows:

<!-- MdxReplaceCode(misc/hello.c) -->

` + "```" + `
foo
` + "```" + `

<!-- MdxToc(misc/*.md) { -->
<!-- } -->

* Another C source in a list follows

	<!-- MdxReplaceCode(misc/world.c) -->

	` + "```" + `
	bar
	` + "```" + `

Next, indented:

* foo

	<!-- MdxReplaceCode(misc/hello.c) -->

		aaa

		bbb

	<!-- -->

		ccc

	* bar

		<!-- MdxReplaceCode(misc/world.c) -->

		~~~c:world.c
		baz
		~~~

Rest

` + "```" + `
foo
bar
` + "```" + `

Lorem ipsum “<!-- MdxLink(misc/foo.md) { -->foo<!-- } -->” dolor “<!-- MdxLink(misc/bar.md) { -->bar<!-- } -->” sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

Yeah.
`

	expectedOutputString := `A C source follows:

<!-- MdxReplaceCode(misc/hello.c) -->

` + "```" + `
#include <stdio.h>

int main (int argc, char** argv) {
	printf("Hello!\n");
}
` + "```" + `

<!-- MdxToc(misc/*.md) { -->
* [Bar ドキュメント](misc/bar.md)
* [misc/foo.md](misc/foo.md)
<!-- } -->

* Another C source in a list follows

	<!-- MdxReplaceCode(misc/world.c) -->

	` + "```" + `
	#include <stdio.h>
	
	int main (int argc, char** argv) {
		printf("World!\n");
	}
	` + "```" + `

Next, indented:

* foo

	<!-- MdxReplaceCode(misc/hello.c) -->

		#include <stdio.h>
		
		int main (int argc, char** argv) {
			printf("Hello!\n");
		}

	<!-- -->

		ccc

	* bar

		<!-- MdxReplaceCode(misc/world.c) -->

		~~~c:world.c
		#include <stdio.h>
		
		int main (int argc, char** argv) {
			printf("World!\n");
		}
		~~~

Rest

` + "```" + `
foo
bar
` + "```" + `

Lorem ipsum “<!-- MdxLink(misc/foo.md) { -->[misc/foo.md](misc/foo.md)<!-- } -->” dolor “<!-- MdxLink(misc/bar.md) { -->[Bar ドキュメント](misc/bar.md)<!-- } -->” sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

Yeah.
`

	input := strings.NewReader(inputString)
	var output = &strings.Builder{}
	err := Preprocess(input, output)
	if err != nil {
		t.Fatal("Error occurred.", err)
	}
	if expectedOutputString != output.String() {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(expectedOutputString, output.String()))
	}
}

func TestMdxOpenFailure(t *testing.T) {
	inputString := `A C source follows:

<!-- MdxReplaceCode(misc/hello.cc) -->

` + "```" + `
foo
` + "```" + `
`

	input := strings.NewReader(inputString)
	var output = &strings.Builder{}
	err := Preprocess(input, output)
	if err == nil || err.Error() != "open misc/hello.cc: no such file or directory" {
		t.Fatal("Error")
	}
}

func TestMdxOpenIndentedFailure(t *testing.T) {
	inputString := `A C source follows:

<!-- MdxReplaceCode(misc/world.cc) -->

	foo
`

	input := strings.NewReader(inputString)
	var output = &strings.Builder{}
	err := Preprocess(input, output)
	if err == nil || err.Error() != "open misc/world.cc: no such file or directory" {
		t.Fatal("Error")
	}
}

func TestBlockMissing(t *testing.T) {
	inputString := `A C source follows:

<!-- MdxReplaceCode(misc/world.cc) -->
`

	input := strings.NewReader(inputString)
	var output = &strings.Builder{}
	err := Preprocess(input, output)
	if err.Error() != "not ground" {
		t.Fatal("Error")
	}
}

func TestIgnoreUnknownCommand(t *testing.T) {
	inputString := `A C source follows:

<!-- MdxHogeHoge(foobar) { -->
foo
bar
<!-- } -->
hoge
fuga
`

	input := strings.NewReader(inputString)
	var output = &strings.Builder{}
	_ = Preprocess(input, output)

	if inputString != output.String() {
		t.Fatalf(`Unmatched:

%s`, diff.LineDiff(inputString, output.String()))
	}
}

func TestRegexpReplaceAll(t *testing.T) {
	line := "foo#bar hoge#fuga"
	re := regexp.MustCompile(`([a-z]+)#([a-z]+)`)
	modified := replaceAllStringSubMatchFunc(re, line, func(a []string) string {
		return a[1] + "@" + a[2]
	})
	if modified != "foo@bar hoge@fuga" {
		t.Fatal("Unmatched")
	}
}
