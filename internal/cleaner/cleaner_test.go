package cleaner

import "testing"

// Эталонные выводы сняты с Node-версии: `node dist/src/main.js clean --stdin`
// (stdin парсится как stdin.ts, keepJSDoc=true, whitelist по умолчанию ["//!"]).
func TestCleanParityWithNode(t *testing.T) {
	c := New([]string{"//!"})
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			"line_and_block",
			"const a = 1; // line\n/* block */\nconst b = 2;",
			"const a = 1; \n\nconst b = 2;",
		},
		{
			"string_with_slashes",
			`const url = "https://example.com"; // real`,
			`const url = "https://example.com"; `,
		},
		{
			"template_literal",
			"const t = `a // not comment ${x}`; // real",
			"const t = `a // not comment ${x}`; ",
		},
		{
			"regex_literal",
			`const re = /\/\* not \*\//g; // real`,
			`const re = /\/\* not \*\//g; `,
		},
		{
			"jsdoc_kept",
			"/** keep me */\nfunction f() {} // remove",
			"/** keep me */\nfunction f() {} ",
		},
		{
			"whitelist_tsignore",
			"// @ts-ignore\nconst x: any = 1; // remove",
			"// @ts-ignore\nconst x: any = 1; ",
		},
		{
			"whitelist_bang",
			"//! keep this\n// remove this",
			"//! keep this\n",
		},
		{
			"ident_boundary",
			"const x = a/*c*/b;",
			"const x = a b;",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := c.Clean("stdin.ts", tc.in, true).Content
			if got != tc.want {
				t.Fatalf("\n in:   %q\n got:  %q\n want: %q", tc.in, got, tc.want)
			}
		})
	}
}

// JSDoc удаляется при keepJSDoc=false (флаг --no-jsdoc).
func TestNoJSDoc(t *testing.T) {
	c := New([]string{"//!"})
	got := c.Clean("stdin.ts", "/** doc */\nconst x = 1;", false).Content
	if got != "\nconst x = 1;" {
		t.Fatalf("got %q", got)
	}
}

// Пустое JSX-выражение `{/* */}` удаляется целиком (вместе со скобками).
func TestJSXEmptyExpressionRemovedWhole(t *testing.T) {
	c := New([]string{"//!"})
	got := c.Clean("App.tsx", "const a = <div>{/* x */}</div>;", true).Content
	if got != "const a = <div></div>;" {
		t.Fatalf("got %q", got)
	}
}

// JSX-выражение с реальным выражением сохраняется, удаляется лишь комментарий.
func TestJSXWithExpressionKeepsBraces(t *testing.T) {
	c := New([]string{"//!"})
	got := c.Clean("App.tsx", "const a = <div>{x /* c */}</div>;", true).Content
	if got != "const a = <div>{x }</div>;" {
		t.Fatalf("got %q", got)
	}
}

// HTML-комментарии вырезаются регуляркой.
func TestHTMLComment(t *testing.T) {
	c := New([]string{"//!"})
	got := c.Clean("page.html", "<!-- c -->\n<div>x</div>", true).Content
	if got != "\n<div>x</div>" {
		t.Fatalf("got %q", got)
	}
}

// Нет комментариев — содержимое не меняется, removed=0.
func TestNoCommentsUnchanged(t *testing.T) {
	c := New([]string{"//!"})
	r := c.Clean("stdin.ts", "const x = 1;", true)
	if r.Removed != 0 || r.Content != "const x = 1;" {
		t.Fatalf("got %+v", r)
	}
}
