// See original Lispy at http://norvig.com/lispy.html

package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strings"
)

// Read a Scheme expression from a string.
func parse(program string) interface{} {
    tokens := tokenize(program)
    s, _ := readFromTokens(tokens)
    return s
}

// Convert a string of characters into a list of tokens.
func tokenize(chars string) []string {
    s := strings.Replace(chars, "(", " ( ", -1)
    s = strings.Replace(s, ")", " ) ", -1)
    s = strings.TrimSpace(s)
    return regexp.MustCompile("\\s+").Split(s, -1)
}

func pop(ts []string) (string, []string) {
    if len(ts) == 0 {
	return "", make([]string, 0)
    } else {
	return ts[0], ts[1:]
    }
}

// Read an expression from a sequence of tokens.
func readFromTokens(tokens []string) (interface{}, []string) {
    if len(tokens) == 0 {
        panic("unexpected EOF while reading")
    }
    var token string
    token, tokens = pop(tokens)
    if "(" == token {
	l := make([]interface{}, 0)
        for tokens[0] != ")" {
	    var s interface{}
	    s, tokens = readFromTokens(tokens)
            l = append(l, s)
	}
        _, tokens = pop(tokens) // pop off ")"
        return l, tokens
    } else if ")" == token {
        panic("unexpected )")
    } else {
        return atom(token), tokens
    }
}

// Numbers become numbers; every other token is a symbol.
func atom(token string) interface{} {
    var n int
    _, err := fmt.Sscan(token, &n)
    if err == nil {
	return n
    } else {
	return token
    }
}

// An environment with some Scheme standard procedures.
func standardEnv() map[string]interface{} {
    env := make(map[string]interface{}, 0)
    env["false"] = false
    env["true"] = true
    env["+"] = func(args []interface{}) interface{} {
	n, ok1 := args[0].(int)
	m, ok2 := args[1].(int)
	if !ok1 || !ok2 {
	    panic("+ needs numbers")
	}
	return n + m
    }
    return env
}

var globalEnv = standardEnv()

// Evaluate an expression in an environment.
func eval(x interface{}, env map[string]interface{}) interface{} {
    if str, ok := isSymbol(x); ok { // variable reference
	return env[str]
    }
    l, ok := isList(x)
    if !ok { // constant literal
	return x
    }
    if len(l) == 0 {
	panic("empty list")
    }
    if str, ok := isSymbol(l[0]); ok {
	switch (str) {
	case "quote": // (quote exp)
	    return l[1]
	case "if": //  (if test conseq alt)
	    test := l[1]
	    conseq := l[2]
	    alt := l[3]
	    r := eval(test, env)
	    if b, ok := isFalse(r); ok && !b {
		return eval(alt, env)
	    } else {
		return eval(conseq, env)
	    }
	case "define": // (define var exp)
	    car := l[1]
	    cdr := l[2]
	    if str, ok = isSymbol(car); ok {
		env[str] = eval(cdr, env)
		return env[str]
	    } else {
		panic("define needs a symbol")
	    }
	default: // (proc arg...)
	    car := eval(l[0], env)
	    proc, ok := car.(func([]interface{})interface{})
	    if !ok {
		panic("not a procedure")
	    }
            args := makeArgs(l[1:], env)
            return proc(args)
	}
    }
    return nil
}

func makeArgs(l []interface{}, env map[string]interface{}) []interface{} {
    args := make([]interface{}, 0)
    for i := 0; i < len(l); i++ {
	args = append(args, eval(l[i], env))
    }
    return args
}

func isSymbol(x interface{}) (string, bool) {
    s, ok := x.(string)
    return s, ok
}

func isList(x interface{}) ([]interface{}, bool) {
    l, ok := x.([]interface{})
    return l, ok
}

func isFalse(x interface{}) (bool, bool) {
    b, ok := x.(bool)
    return b, ok
}

// A prompt-read-eval-print loop.
func repl() {
    prompt := "lis.py> "
    in := bufio.NewReader(os.Stdin)
    for {
	fmt.Printf("%s", prompt)
	line, _ := in.ReadString('\n')
	val := eval(parse(line), globalEnv)
        if val != nil {
            fmt.Printf("%s\n", schemestr(val))
	}
    }
}

// Convert a Python object back into a Scheme-readable string.
func schemestr(exp interface{}) string {
    if l, ok := isList(exp); ok {
	s := make([]string, 0)
	for i := 0; i < len(l); i++ {
	    s = append(s, schemestr(l[i]))
	}
        return "(" + strings.Join(s, " ") + ")"
    } else {
	return fmt.Sprintf("%d", exp)
    }
}

func dump(s string) {
    d(parse(s))
}

func d(a interface{}) {
    fmt.Printf("%s\n", a)
}

func e(s string) {
    d(eval(parse(s), globalEnv))
}

func t() {
    dump("1")
    dump("a")
    dump("()")
    dump("(1)")
    dump("(1 2)")
    dump("(1 (2))")
    str, ok := isSymbol("a")
    fmt.Printf("%s %s\n", str, ok)
    str, ok = isSymbol(1)
    fmt.Printf("%s %s\n", str, ok)
    e("(quote ())")
    e("(if false 1 2)")
    e("(if true 1 2)")
    e("(define a 10)")
    e("a")
    e("(+ 1 2)")
}

func main() {
    repl()
}

