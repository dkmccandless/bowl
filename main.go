package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for fmt.Print("> "); scanner.Scan(); fmt.Print("> ") {
		REPL(scanner.Text())
	}
}

func REPL(in string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	exp, err := Parse(in)
	if err != nil {
		panic(err)
	}
	fmt.Println(Eval(Expression(exp), GlobalEnv))
}

// One of: int, float64, string, []Expression
type Expression Value

// Elements in a Stack may not be empty. Empty string signals EOF.
type Stack []string

func (s *Stack) Shift() string {
	if len(*s) == 0 {
		return ""
	}
	t := (*s)[0]
	*s = (*s)[1:]
	return t
}

func (s *Stack) Peek() string {
	if len(*s) == 0 {
		return ""
	}
	t := (*s)[0]
	return t
}

func Tokenize(program string) (tokens []string) {
	program = strings.Replace(program, "(", " ( ", -1)
	program = strings.Replace(program, ")", " ) ", -1)
	return strings.Fields(program)
}

func ReadFromTokens(tokens *Stack) (Value, error) {
	t := tokens.Shift()
	if t == "" {
		return nil, errors.New("empty expression")
	}
	if t == ")" {
		return nil, errors.New("unexpected closing parenthesis")
	}
	if t == "(" {
		var e Value
		for tokens.Peek() != ")" {
			exp, err := ReadFromTokens(tokens)
			if err != nil {
				return nil, err
			}
			e = schemeAppend(e, exp)
		}
		_ = tokens.Shift()
		return e, nil
	}
	if n, err := strconv.Atoi(t); err == nil {
		return n, nil
	}
	if n, err := strconv.ParseFloat(t, 64); err == nil {
		return n, nil
	}
	return t, nil
}

func Parse(program string) (Value, error) {
	stack := Stack(Tokenize(program))
	v, err := ReadFromTokens(&stack)
	if err != nil {
		return v, err
	}
	if stack.Peek() != "" {
		return v, errors.New("trailing garbage")
	}
	return v, nil
}

type Pair struct{ a, d Value }

func (p Pair) String() string {
	//
}

// One of: bool, int, float64, string, Pair, func(...Value) Value
type Value interface{}

func Eval(expr Expression, env Env) Value {
	switch expr := expr.(type) {
	case int:
		return expr
	case float64:
		return expr
	case string:
		v, ok := env[expr]
		if !ok {
			panic("Unknown variable")
		}
		return v
	case Pair:
		switch op, _ := car(expr).(string); op {
		case "if":
			return EvalIf(cadr(expr), caddr(expr), cadddr(expr), env)
		case "define":
			return EvalDefine(cadr(expr), caddr(expr), env)
		case "quote":
			return cadr(Value(expr))
		default:
			// (x a b c)
			f := Eval(car(expr), env)
			var values []Value
			for e := cdr(expr); e != nil; e = cdr(e) {
				values = append(values, Eval(car(e), env))
			}
			return Apply(f, values)
		}
	}
	panic("eval: bug in the interpreter")
}

func EvalIf(test, consequent, alt Expression, env Env) Value {
	if Eval(test, env).(bool) {
		return Eval(consequent, env)
	} else {
		return Eval(alt, env)
	}
}

func EvalDefine(v, exp Expression, env Env) Value {
	env[v.(string)] = Eval(exp, env)
	return "ok"
}

func Apply(f Value, args []Value) Value {
	// f is func() Value, func(Value) Value, func(Value, Value) Value, etc => func(*[Value]) Value
	values := make([]reflect.Value, len(args))
	for i := range args {
		values[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(f).Call(values)[0].Interface().(Value)
}

type Env map[string]Value

var GlobalEnv = Env{
	"true":  true,
	"false": false,
	"car":   car,
	"cdr":   cdr,
	"cons":  cons,
	"=":     equal,
	"<":     lessthan,
	">":     greaterthan,
	"+":     add,
	"-":     sub,
	"*":     mult,
	"/":     div,
	// "abs":    abs,
	"append": schemeAppend,
}

// (car (cons 4 5)) => 4
func car(p Value) Value {
	return p.(Pair).a
}

func cdr(p Value) Value {
	return p.(Pair).d
}

func cadr(p Value) Value   { return car(cdr(p)) }
func caddr(p Value) Value  { return cadr(cdr(p)) }
func cadddr(p Value) Value { return caddr(cdr(p)) }

func cons(a, d Value) Value {
	return Pair{a, d}
}

func equal(a, b Value) Value {
	return reflect.DeepEqual(a, b)
}

func lessthan(a, b Value) Value {
	switch a := a.(type) {
	case int:
		return a < b.(int)
	case float64:
		return a < b.(float64)
	}
	panic(fmt.Sprintf("add: mismatched Value types %T and %T", a, b))
}

func greaterthan(a, b Value) Value {
	switch a := a.(type) {
	case int:
		return a > b.(int)
	case float64:
		return a > b.(float64)
	}
	panic(fmt.Sprintf("add: mismatched Value types %T and %T", a, b))
}

func add(a, b Value) Value {
	switch a := a.(type) {
	case int:
		return a + b.(int)
	case float64:
		return a + b.(float64)
	}
	panic(fmt.Sprintf("add: mismatched Value types %T and %T", a, b))
}

func sub(a, b Value) Value {
	switch a := a.(type) {
	case int:
		return a - b.(int)
	case float64:
		return a - b.(float64)
	}
	panic(fmt.Sprintf("sub: mismatched Value types %T and %T", a, b))
}

func mult(a, b Value) Value {
	switch a := a.(type) {
	case int:
		return a * b.(int)
	case float64:
		return a * b.(float64)
	}
	panic(fmt.Sprintf("mult: mismatched Value types %T and %T", a, b))
}

func div(a, b Value) Value {
	switch a := a.(type) {
	case int:
		return a / b.(int)
	case float64:
		return a / b.(float64)
	}
	panic(fmt.Sprintf("div: mismatched Value types %T and %T", a, b))
}

// (define nil (quote ()))

//                        a               b
//         ------------------------------ -
// (append (cons 1 (cons 2 (cons 3 nil))) 4) => (1 2 3 4)
// (append (cons 2 (cons 3 nil)) 4) => (2 3 4)
// (append (cons 3 nil) 4) => (3 4)
// (append nil 4) => (4)
func schemeAppend(a, b Value) Value {
	if a == nil {
		return Pair{b, nil}
	}
	return cons(car(a), schemeAppend(cdr(a), b))
}

/*
(define (append a b)
  (if (= a nil)
    (cons b nil)
    (cons (car a) (append (cdr a) b)))
*/
