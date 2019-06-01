package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"strconv"
	"testing"
)

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got = %q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got = %T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got = %s.", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got = %s.", name, letStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il is not *ast.IntegerLiteral. got = %T.", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value is not %d. got = %d.", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral is not %d. got = %s.", value, integ.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp is not *ast.Identifier. got = %T.", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value is not %s. got = %s.", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral is not %s. got = %s.", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	b, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp is not *ast.Boolean. got = %T.", exp)
		return false
	}

	if b.Value != value {
		t.Errorf("b.Value is not %t. got = %t.", value, b.Value)
		return false
	}

	if b.TokenLiteral() != strconv.FormatBool(value) {
		t.Errorf("b.TokenLiteral is not %s. got = %s.", strconv.FormatBool(value), b.TokenLiteral())
		return false
	}

	return true
}

// リテラルの型に応じて exp の型と値チェックを行うテスト
func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}

	t.Errorf("type of exp is not handled. got = %T.", exp)
	return false
}

func testPrefixExpression(t *testing.T, exp ast.Expression, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.PrefixExpression)

	if !ok {
		t.Errorf("exp is not ast.PrefixExpression. got = %T(%s).", exp, exp)
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got = %q.", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)

	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got = %T(%s).", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got = %q.", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

//----------------------------------------
// ユーティリティ
//----------------------------------------
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func InitializeTest(t *testing.T, input string, expectLength int) *ast.Program {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != expectLength {
		t.Fatalf("program has not enough statements. got = %d", len(program.Statements))
	}

	return program
}

//----------------------------------------
// 文のテスト
//----------------------------------------
func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		program := InitializeTest(t, tt.input, 1)

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return 10;", 10},
		{"return 993322;", 993322},
	}

	for _, tt := range tests {
		program := InitializeTest(t, tt.input, 1)
		returnStmt, ok := program.Statements[0].(*ast.ReturnStatement)

		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got = %T", program.Statements[0])
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return'. got = %q", returnStmt.TokenLiteral())
		}

		val := returnStmt.ReturnValue
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

//----------------------------------------
// 式のテスト
//----------------------------------------
func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression is not *ast.Identifier. got = %T.", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("ident.Value is not %s. got = %s.", "foobar", ident.Value)
	}

	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral is not %s. got = %s.", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression is not *ast.IntegerLiteral. got = %T.", stmt.Expression)
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value is not %d. got = %d.", 5, literal.Value)
	}

	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral is not %s. got = %s.", "5", literal.TokenLiteral())
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	program := InitializeTest(t, input, 1)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)

	if !ok {
		t.Fatalf("exp is not *ast.StringLiteral. got = %T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value is not %q. got = %q", "hello world", literal.Value)
	}
}

func TestBooleanExpression(t *testing.T) {
	input := "true;"

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	b, ok := stmt.Expression.(*ast.Boolean)
	if !ok {
		t.Fatalf("expression is not *ast.Boolean. got = %T.", stmt.Expression)
	}

	if b.Value != true {
		t.Errorf("b.Value is not %t. got = %t.", true, b.Value)
	}

	if b.TokenLiteral() != "true" {
		t.Errorf("b.TokenLiteral is not %s. got = %s.", "true", b.TokenLiteral())
	}
}

func TestParsingArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp is not ast.ArrayLiteral. got = %T.", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) is not 3. got = %d", len(array.Elements))
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "myArray[1 + 2]"

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)

	if !ok {
		t.Fatalf("exp is not *ast.IndexExpression. got = %T", stmt.Expression)
	}

	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, indexExp.Index, 1, "+", 2) {
		return
	}
}

func TestParsingHashLiteralsStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got = %T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got = %d", len(hash.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got = %T", key)
		}

		expectedValue := expected[literal.String()]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingEmptyHashLiteral(t *testing.T) {
	input := "{}"

	program := InitializeTest(t, input, 1)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got = %T", stmt.Expression)
	}

	if len(hash.Pairs) != 0 {
		t.Errorf("hash.Pairs has wrong length. got = %d", len(hash.Pairs))
	}
}

func TestParsingHashLiteralsWithExpressions(t *testing.T) {
	input := `{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`

	program := InitializeTest(t, input, 1)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got = %T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got = %d", len(hash.Pairs))
	}

	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got = %T", key)
		}

		testFunc, ok := tests[literal.String()]
		if !ok {
			t.Errorf("No test function for key %q found.", literal.String())
			continue
		}

		testFunc(value)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		program := InitializeTest(t, tt.input, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
				program.Statements[0])
		}

		if !testPrefixExpression(t, stmt.Expression, tt.operator, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 6;", 5, "+", 6},
		{"5 - 6;", 5, "-", 6},
		{"5 * 10;", 5, "*", 10},
		{"5 / 6;", 5, "/", 6},
		{"8 > 6;", 8, ">", 6},
		{"5 < 6;", 5, "<", 6},
		{"5 == 6;", 5, "==", 6},
		{"5 != 6;", 5, "!=", 6},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, tt := range infixTests {
		program := InitializeTest(t, tt.input, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
				program.Statements[0])
		}

		if !(testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue)) {
			return
		}
	}
}

func TestParseOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, tt := range tests {
		program := InitializeTest(t, tt.input, 1)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected = %q, got = %q.", tt.expected, actual)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`
	expectLength := 1
	program := InitializeTest(t, input, expectLength)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got = %T",
			stmt.Expression)
	}

	// ここで infix を使うのは単に input が中置構文だから?
	if !(testInfixExpression(t, exp.Condition, "x", "<", "y")) {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got = %d\n", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Consequence.Statement[0] is not ast.ExpressionStatement. got = %T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative.Statements was not nil. got = %+v", exp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`
	expectLength := 1
	program := InitializeTest(t, input, expectLength)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got = %T",
			stmt.Expression)
	}

	// ここで infix を使うのは単に input が中置構文だから?
	if !(testInfixExpression(t, exp.Condition, "x", "<", "y")) {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got = %d\n", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Consequence.Statement[0] is not ast.ExpressionStatement. got = %T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative == nil {
		t.Errorf("exp.Alternative.Statements is nil.")
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("alternative is not 1 statements. got = %d\n", len(exp.Alternative.Statements))
	}

	alt, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Alternative.Statement[0] is not ast.ExpressionStatement. got = %T", exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alt.Expression, "y") {
		return
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T.", program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got = %T.", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters is wrong. expected = %d, got = %d\n", 2, len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statements. got = %d\n", len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function.Body.Statement[0] is not ast.ExpressionStatement. got = %T.",
			function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		program := InitializeTest(t, tt.input, 1)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("parameters length is wrong. expected = %d, got = %d\n",
				len(tt.expectedParams), len(function.Parameters))
		}

		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	program := InitializeTest(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got = %T.", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got = %T.", stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("arguments length is wrong. expected = %d, got = %d.\n", 3, len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}
