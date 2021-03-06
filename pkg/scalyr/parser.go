// Copyright 2019 Adknown Inc. All rights reserved.
// Created:  2019-08-11
// Author:   matt
// Project:  grafana-scalyr-datasource-plugin

package scalyr

import (
	"fmt"
	"github.com/golang-collections/collections/stack"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

type Token struct {
	Variable string      //the respresentative variable for the token
	Data     interface{} //the data that the token holds
}

type ParseVar struct {
	Id         string
	Filter     string
	Query      *TimeseriesQuery
	Response   *TimeseriesQueryResponse
	ConstValue float64 //this is just for the constant values
}

const (
	PrecedenceLevel1   = 1
	PrecedenceLevel2   = 2
	FloatCutoff        = 0.0001
	DivideByZeroResult = -1
	ValueIndex         = 0
	TimestampIndex     = 1
	ReplacePrefix      = "var"
)

/**
 * @param        $expression
 * @param        $start
 * @param        $end
 * @param        $buckets
 * @param string $fullVariableExpression
 *
 * @param        $useNumeric
 *
 * @return array
 * @throws \Adknown\ProxyScalyr\Scalyr\Request\Exception\BadBucketsException
 */
func ParseComplexExpression(expression string, start string, end string, buckets int, fullVariableExpression *string, useNumeric bool) ([]*ParseVar, error) {
	/*
	 * Match graph public static function calls
	 * Part 1
	 * (count|rate|mean|min|max|sumPerSecond|median|fraction|p|p\\[\\d+\\]|p\\d+) match any of the graph function
	 * keywords (One of count, rate, mean, min, max, sumPerSecond, median, fraction, p, p[##] or p##)
	 * p\\[\\d+\\] matches p the left [ and one or more digits the right ]
	 * p\\d+ matches p and one or more digits
	 * Part 2
	 * \\((.*?)\\) matches left ( and everything inside up until right ) in a non greedy way
	 */
	functionRegex := `(?P<function>count|rate|mean|min|max|sumPerSecond|median|p|p\d+)\(((?P<field>[a-zA-Z]+) where )?(?P<filter>.*?)\)`

	/*
	 * Match any constant
	 * \\b  boundry on the left side
	 * \\d+ one or more digits
	 * (?:\.\d+)? Non capturing group to check if a decimal is there
	 * \\b  boundry on the right side
	 */
	numRegex := `\b\d+(?:\.\d+)?\b`

	regex := regexp.MustCompile(functionRegex + "|" + numRegex)

	//Replace all API calls with a placeholder with var prefix
	varCount := 0
	varArray := make([]*ParseVar, 0)

	//Replace all constants with a placeholder with var prefix
	//Match all constants
	for {
		replaceString := fmt.Sprintf("%s%d", ReplacePrefix, varCount)
		submatch := regex.FindStringSubmatch(expression)
		if len(submatch) == 0 {
			break
		}
		val, err := strconv.ParseFloat(submatch[0], 64)
		if err != nil {
			matches := make(map[string]string)
			for i, name := range regex.SubexpNames() {
				if i != 0 && name != "" {
					matches[name] = submatch[i]
				}
			}
			filter := matches["filter"]
			graphFunction := matches["function"]

			varArray = append(varArray, &ParseVar{
				Id:     replaceString,
				Filter: submatch[0],
				Query: &TimeseriesQuery{
					Filter:    filter,
					Buckets:   buckets,
					Function:  graphFunction,
					StartTime: start,
					EndTime:   end,
					Priority:  "low",
				},
				Response:   nil,
				ConstValue: 0,
			})
			expression = strings.Replace(expression, submatch[0], replaceString, 1)
		} else {
			varArray = append(varArray, &ParseVar{
				Id:         replaceString,
				Filter:     submatch[0],
				Query:      nil,
				Response:   nil,
				ConstValue: val,
			})
			//If you try to replace like above, and you have a constant that is less then or equal to the number of 'vars'
			//it will just replace in the 'var'
			//i.e. var0 + var 1 + 0 will turn into varvar1 + var1 + 0
			thisNumRegex := regexp.MustCompile(`\b` + submatch[0] + `\b`)
			expression = thisNumRegex.ReplaceAllString(expression, replaceString)
		}

		varCount++
	}

	*fullVariableExpression = expression

	return varArray, nil
}

//convertInfixNotationToReversePolishNotation changes an infix notation equation in to
//an array in reverse polish notation.
//@see http://andreinc.net/2010/10/05/converting-infix-to-rpn-shunting-yard-algorithm/
//@see https://en.wikipedia.org/wiki/Reverse_Polish_notation
func convertInfixNotationToReversePolishNotation(inputExpressions []string) (*queue.Queue, error) {
	curStack := stack.New()
	output := queue.New(0)
	var stackToken Token

	//Change iterate though infix and change the order to RPN
	for _, tokenStr := range inputExpressions {
		token := Token{
			Variable: tokenStr,
			Data:     nil,
		}
		if isOperator(token) {
			for curStack.Len() > 0 && isOperator(curStack.Peek()) {
				if precedenceCompare(token, curStack.Peek()) <= 0 {
					err := output.Put(curStack.Pop())
					if err != nil {
						return nil, err
					}
					continue
				}
				break
			}
			curStack.Push(token)
		} else {
			if token.Variable == "(" {
				curStack.Push(token)
			} else {
				if token.Variable == ")" {
					for curStack.Len() > 0 {
						if i, ok := curStack.Peek().(Token); !ok {
							return nil, errors.New(fmt.Sprintf("Non Token value found in stack: %v", curStack.Peek()))
						} else {
							stackToken = i
						}
						if stackToken.Variable == "(" {
							break
						}
						err := output.Put(curStack.Pop())
						if err != nil {
							return nil, err
						}
					}
					curStack.Pop()
				} else {
					err := output.Put(token)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	//Empty stack
	for curStack.Len() > 0 {
		err := output.Put(curStack.Pop())
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

/**
 * isOperator Figures out whether the input character is an operator
 *
 * @param $inputOperator
 *
 * @return bool whether the input is an operator
 */
func isOperator(token interface{}) bool {
	if inputOperator, ok := token.(Token); !ok {
		return false
	} else {
		switch inputOperator.Variable {
		case "+":
			fallthrough
		case "-":
			fallthrough
		case "*":
			fallthrough
		case "/":
			return true
		default:
			return false
		}
	}
}

//precedenceCompare Compares the precedence for two operators so they can follow BEDMAS
func precedenceCompare(t1 interface{}, t2 interface{}) int {
	var firstToken, secondToken Token
	if i, ok := t1.(Token); !ok {
		panic("t1 was not a Token")
	} else {
		firstToken = i
	}
	if i, ok := t2.(Token); !ok {
		panic("t2 was not a Token")
	} else {
		secondToken = i
	}

	//Higher precedence operators
	higherPrecedence := map[string]int{
		"*": 0,
		"/": 0,
	}

	//Set the first precedence
	firstPrecedence := PrecedenceLevel1
	if _, ok := higherPrecedence[firstToken.Variable]; ok {
		firstPrecedence = PrecedenceLevel2
	}

	//Set the second precedence
	secondPrecedence := PrecedenceLevel1
	if _, ok := higherPrecedence[secondToken.Variable]; ok {
		secondPrecedence = PrecedenceLevel2
	}

	return firstPrecedence - secondPrecedence
}

//NewEvaluateExpression runs through all the operators/operands and
//calls the Scalyr query requests. It then applies the math operations on the results of the Scalyr queries.
func NewEvaluateExpression(expression string, varArray []*ParseVar) (*TimeseriesQueryResponse, error) {
	totalExecutionTime := 0
	for _, str := range []string{" ", "v", "a", "r"} {
		expression = strings.ReplaceAll(expression, str, "")
	}

	rpnExpression, err := convertInfixNotationToReversePolishNotation(strings.Split(expression, ""))
	if err != nil {
		return nil, errors.Wrap(err, "Err while evaluating an expression and trying to convert from infix notation to reverse polish notation")
	}

	curStack := stack.New()
	var token Token
	for !rpnExpression.Empty() {
		queueEl, err := rpnExpression.Get(1)
		if err != nil {
			return nil, errors.Wrap(err, "Err while evaluating expression and trying to 'Get' on rpn expression quuee")
		}
		if val, ok := queueEl[0].(Token); !ok {
			panic(err)
		} else {
			token = val
		}

		if isOperator(token) {
			//Operator
			result := performOperation(token, curStack.Pop(), curStack.Pop())
			curStack.Push(result)
		} else {
			//Operand
			idx, err := strconv.Atoi(token.Variable)
			if err != nil {
				panic(err)
			}
			curStack.Push(varArray[idx])
			if varArray[idx].Response != nil {
				totalExecutionTime += varArray[idx].Response.ExecutionTime
			}
			continue
		}
	}

	var finalResult *ParseVar
	if val, ok := curStack.Pop().(*ParseVar); !ok {
		panic("Final result of evaluated expression was somehow not a ParseVar Pointer")
	} else {
		finalResult = val
	}

	finalResponse := &TimeseriesQueryResponse{
		Status:        "success",
		Results:       finalResult.Response.Results,
		ExecutionTime: totalExecutionTime,
		Message:       "",
	}
	return finalResponse, nil
}

/**
 * performOperation calls the appropriate equation logic given the operator
 *
 * @param string    $operator  operator  of the equation
 * @param NumericResponse|int $firstVar  first operand of the equation
 * @param NumericResponse|int $secondVar second operand of the equation
 *
 * @return array|float|int|string a single object representing the result of the expression
 */
func performOperation(op interface{}, v1 interface{}, v2 interface{}) *ParseVar {
	var operator Token
	var firstVar, secondVar *ParseVar
	if i, ok := op.(Token); !ok {
		panic("in performOperation(), op was not a Token")
	} else {
		operator = i
	}
	if i, ok := v1.(*ParseVar); !ok {
		panic("in performOperation(), v1 was not a ParseVar pointer")
	} else {
		firstVar = i
	}
	if i, ok := v2.(*ParseVar); !ok {
		panic("in performOperation(), v2 was not a ParseVar pointer")
	} else {
		secondVar = i
	}
	switch operator.Variable {
	case "+":
		return logic(firstVar, secondVar, add)
	case "-":
		return logic(firstVar, secondVar, sub)

	case "/":
		return logic(firstVar, secondVar, div)

	case "*":
		return logic(firstVar, secondVar, mul)
	default:
		return nil
	}
}

func add(a float64, b float64) float64 {
	return a + b
}

func sub(a float64, b float64) float64 {
	return a - b
}

func div(a float64, b float64) float64 {
	if b == 0 {
		return DivideByZeroResult
	}
	if float64(b) < FloatCutoff {
		return DivideByZeroResult
	}

	return a / b
}

func mul(a float64, b float64) float64 {
	return a * b
}

/**
 * addLogic all logic for addition of the 4 possible query cases
 *
 * @param          $firstVar  NumericResponse|int first operand of the equation
 * @param          $secondVar NumericResponse|int second operand of the equation
 * @param callback $op        second operand of the equation
 *
 * @return NumericResponse|int a single object representing the result of the expression
 */
func logic(firstVar *ParseVar, secondVar *ParseVar, op func(float64, float64) float64) *ParseVar {
	newFilter := fmt.Sprintf("%s %s %s", firstVar.Filter, "*", secondVar.Filter)
	//Both are queries
	if firstVar.Response != nil && secondVar.Response != nil {
		//Prepare response
		response := &ParseVar{
			firstVar.Id,
			newFilter,
			firstVar.Query,
			firstVar.Response,
			firstVar.ConstValue,
		}

		for idx, val := range firstVar.Response.Results[0].Values {
			response.Response.Results[0].Values[idx] = op(secondVar.Response.Results[0].Values[idx], val)
		}

		return response
	}

	//Constant and query or constant and constant
	//First is constant
	if firstVar.Response == nil && secondVar.Response != nil {
		//Prepare response
		response := &ParseVar{
			Id:         secondVar.Id,
			Filter:     newFilter,
			Query:      secondVar.Query,
			Response:   secondVar.Response,
			ConstValue: secondVar.ConstValue,
		}
		for idx, val := range secondVar.Response.Results[0].Values {
			response.Response.Results[0].Values[idx] = op(val, float64(firstVar.ConstValue))
		}

		return response
	}

	//Second is constant
	if firstVar.Response != nil && secondVar.Response == nil {
		//Prepare response
		response := &ParseVar{
			firstVar.Id,
			newFilter,
			firstVar.Query,
			firstVar.Response,
			firstVar.ConstValue,
		}

		for idx, val := range firstVar.Response.Results[0].Values {
			response.Response.Results[0].Values[idx] = op(float64(secondVar.ConstValue), val)
		}

		return response
	}

	//Both are constant
	val := op(float64(secondVar.ConstValue), float64(firstVar.ConstValue))
	return &ParseVar{
		Id:         firstVar.Id,
		Filter:     newFilter,
		Query:      nil,
		Response:   nil,
		ConstValue: val,
	}
}
