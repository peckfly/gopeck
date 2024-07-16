package interpreter

import (
	"errors"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"go.uber.org/zap"
	"math/rand"
	"reflect"
	"strconv"
)

const (
	packName     = "pack"
	packageConst = "package "
)

var (
	evalScripCodePanicError    = errors.New("error: eval script code panic fail")
	evalScriptCodeError        = errors.New("error: eval script code fail")
	evalScriptCodeUserLibError = errors.New("error: eval script code use lib fail")
	evalFunctionError          = errors.New("error: eval function fail")
)

type (
	EvalInterpreter struct {
		executor any
	}
	option struct {
		funcName string
	}
	Option func(*option)
)

// NewEvalInterpreter create new EvalInterpreter
func NewEvalInterpreter(sourceScript string, opts ...Option) (*EvalInterpreter, error) {
	opt := &option{
		funcName: "Check",
	}
	for _, o := range opts {
		o(opt)
	}
	inter := interp.New(interp.Options{})
	err := inter.Use(stdlib.Symbols)
	if err != nil {
		log.Error(evalScriptCodeUserLibError.Error(), zap.Error(err))
		return nil, evalScriptCodeError
	}
	randomPack := packName + strconv.Itoa(rand.Int())
	sourceCode := packageConst + randomPack + "\n" + sourceScript + "\n"
	_, err = inter.Eval(sourceCode)
	if err != nil {
		log.Error(evalScriptCodeError.Error(), zap.Error(err))
		return nil, evalScriptCodeError
	}
	var v reflect.Value
	v, err = inter.Eval(randomPack + "." + opt.funcName)
	if err != nil {
		log.Error(evalFunctionError.Error(), zap.Error(err))
		return nil, evalFunctionError
	}
	return &EvalInterpreter{executor: v.Interface()}, nil
}

// ExecuteScript execute script by responseBody and sourceScript
// script must be have only one function with name Check
// script function must return string
// script function parameter must be map[string]any
func (ev *EvalInterpreter) ExecuteScript(f func(executor any)) (err error) {
	defer func() {
		if p := recover(); p != nil {
			log.ErrorStack(p)
			log.Error(evalScripCodePanicError.Error())
			err = evalScripCodePanicError
			return
		}
	}()
	f(ev.executor)
	return nil
}

func WithFuncName(name string) Option {
	return func(opt *option) {
		opt.funcName = name
	}
}
