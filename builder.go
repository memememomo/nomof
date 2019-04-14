package nomof

import (
	"fmt"
	"strings"
)

// クエリ式
// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/developerguide/Query.html#Query.KeyConditionExpressions

// 比較演算子および関数リファレンス
// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html

// https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/developerguide/DynamoDBMapper.DataTypes.html
type DynamoAttributeType string

const (
	S    DynamoAttributeType = "S"
	SS                       = "SS"
	N                        = "N"
	NS                       = "NS"
	B                        = "B"
	BS                       = "BS"
	BOOL                     = "BOOL"
	NULL                     = "NULL"
	L                        = "L"
	M                        = "M"
)

type Operator string

const (
	EQ Operator = "="
	NE          = "<>"
	LT          = "<"
	LE          = "<="
	GT          = ">"
	GE          = ">="
)

type Builder struct {
	Expr []string
	Arg  []interface{}
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (f *Builder) Op(path string, op Operator, arg interface{}) *Builder {
	return f.Append(
		fmt.Sprintf("'%s' %s ?", path, op),
		[]interface{}{arg},
	)
}

func (f *Builder) Equal(path string, arg interface{}) *Builder {
	return f.Op(path, EQ, arg)
}

func (f *Builder) Between(path string, arg1 interface{}, arg2 interface{}) *Builder {
	return f.Append(
		fmt.Sprintf("'%s' BETWEEN ? AND ?", path),
		[]interface{}{arg1, arg2},
	)
}

func (f *Builder) In(path string, args ...interface{}) *Builder {
	binds := strings.Join(strings.Split(strings.Repeat("?", len(args)), ""), ",")
	return f.Append(
		fmt.Sprintf("'%s' IN (%s)", path, binds),
		args,
	)
}

func (f *Builder) AttributeExists(path string) *Builder {
	return f.appendExpr(fmt.Sprintf("attribute_exists('%s')", path))
}

func (f *Builder) AttributeNotExists(path string) *Builder {
	return f.appendExpr(fmt.Sprintf("attribute_not_exists('%s')", path))
}

func (f *Builder) AttributeType(path string, t DynamoAttributeType) *Builder {
	return f.Append(
		fmt.Sprintf("attribute_type('%s', ?)", path),
		[]interface{}{t},
	)
}

func (f *Builder) BeginsWith(path string, arg interface{}) *Builder {
	return f.Append(
		fmt.Sprintf("begins_with('%s', ?)", path),
		[]interface{}{arg},
	)
}

func (f *Builder) Contains(path string, arg interface{}) *Builder {
	return f.Append(
		fmt.Sprintf("contains('%s', ?)", path),
		[]interface{}{arg},
	)
}

func (f *Builder) Size(path string) *Builder {
	return f.appendExpr(fmt.Sprintf("size('%s')", path))
}

func (f *Builder) generateExps() []string {
	var exps []string
	for _, e := range f.Expr {
		exps = append(exps, fmt.Sprintf("(%s)", e))
	}
	return exps
}

func (f *Builder) JoinAnd() string {
	if len(f.Expr) == 1 {
		return f.Expr[0]
	}
	return strings.Join(f.generateExps(), " AND ")
}

func (f *Builder) JoinOr() string {
	return strings.Join(f.generateExps(), " OR ")
}

func (f *Builder) HasFilter() bool {
	return len(f.Expr) > 0
}

func (f *Builder) Append(expr string, arg []interface{}) *Builder {
	f.appendExpr(expr)
	f.appendArg(arg...)
	return f
}

func (f *Builder) appendExpr(expr string) *Builder {
	f.Expr = append(f.Expr, expr)
	return f
}

func (f *Builder) appendArg(arg ...interface{}) *Builder {
	f.Arg = append(f.Arg, arg...)
	return f
}
