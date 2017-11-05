package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for fmt.Print("> "); scanner.Scan(); fmt.Print("> ") {
		EP(scanner.Text())
	}
}

func EP(in string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	exp, err := Parse(in)
	if err != nil {
		panic(err)
	}
	fmt.Println(Eval(exp, GlobalEnv))
}

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
		return nil, errors.New("empty expession")
	}
	if t == ")" {
		return nil, errors.New("unexpected closing parenthesis")
	}
	if t == "(" {
		var e Value = snil
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
	return fmt.Sprintf("(%v)", p.string())
}

func (p Pair) string() string {
	switch d := p.d.(type) {
	case Pair:
		return fmt.Sprintf("%v %v", p.a, d.string())
	case schemeNil:
		return fmt.Sprintf("%v", p.a)
	default:
		return fmt.Sprintf("%v . %v", p.a, d)
	}
}

// One of: bool, int, float64, string, Pair, func(...Value) Value
type Value interface{}

func Eval(exp Value, env Env) Value {
	switch exp := exp.(type) {
	case int:
		return exp
	case float64:
		return exp
	case string:
		v, ok := env[exp]
		if !ok {
			panic("Unknown variable")
		}
		return v
	case Pair:
		switch op, _ := car(exp).(string); op {
		case "if":
			return EvalIf(cadr(exp), caddr(exp), cadddr(exp), env)
		case "define":
			return EvalDefine(cadr(exp), caddr(exp), env)
		case "quote":
			return cadr(Value(exp))
		default:
			// (x a b c)
			f := Eval(car(exp), env)
			var values []Value
			for e := cdr(exp); e != snil; e = cdr(e) {
				values = append(values, Eval(car(e), env))
			}
			return Apply(f, values)
		}
	}
	panic("eval: bug in the interpreter")
}

func EvalIf(test, consequent, alt Value, env Env) Value {
	if Eval(test, env).(bool) {
		return Eval(consequent, env)
	} else {
		return Eval(alt, env)
	}
}

func EvalDefine(v, exp Value, env Env) Value {
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

type schemeNil struct{}

func (s schemeNil) String() string { return "()" }

var snil schemeNil

var GlobalEnv = Env{
	"true":  true,
	"false": false,
	"nil":   snil,
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
	if a == snil {
		return Pair{b, snil}
	}
	return cons(car(a), schemeAppend(cdr(a), b))
}

/*
(define (append a b)
  (if (= a nil)
    (cons b nil)
    (cons (car a) (append (cdr a) b)))
*/
