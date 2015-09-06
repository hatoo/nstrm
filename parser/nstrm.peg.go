package parser

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruletop
	rulebody
	ruleexpr
	rulee0
	rulee01
	rulee1
	rulee2
	rulee3
	rulee4
	rulevalue
	ruleidentifer
	ruleidentifer_prepare
	ruleidentifer_argment
	rulebind
	rulerefvariable
	rulefuncall
	rulearray
	ruleblock
	ruleifexpr
	rulewhileexpr
	rulewait
	ruleemit
	rulefloating
	ruleinteger
	rulestringliteral
	rulesp
	rulews
	rulecomment
	ruleperiod
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	rulePegText
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27
	ruleAction28
	ruleAction29
	ruleAction30
	ruleAction31
	ruleAction32
	ruleAction33
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38
	ruleAction39
	ruleAction40
	ruleAction41
	ruleAction42
	ruleAction43
	ruleAction44
	ruleminus
	ruleAction45
	ruleAction46
	ruleAction47

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"top",
	"body",
	"expr",
	"e0",
	"e01",
	"e1",
	"e2",
	"e3",
	"e4",
	"value",
	"identifer",
	"identifer_prepare",
	"identifer_argment",
	"bind",
	"refvariable",
	"funcall",
	"array",
	"block",
	"ifexpr",
	"whileexpr",
	"wait",
	"emit",
	"floating",
	"integer",
	"stringliteral",
	"sp",
	"ws",
	"comment",
	"period",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"PegText",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",
	"Action28",
	"Action29",
	"Action30",
	"Action31",
	"Action32",
	"Action33",
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",
	"Action39",
	"Action40",
	"Action41",
	"Action42",
	"Action43",
	"Action44",
	"minus",
	"Action45",
	"Action46",
	"Action47",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	pegRule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens16) PreOrder() (<-chan state16, [][]token16) {
	s, ordered := make(chan state16, 6), t.Order()
	go func() {
		var states [8]state16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens16) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token16{pegRule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token32{pegRule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type Nstrm struct {
	MyParser

	Buffer string
	buffer []rune
	rules  [80]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *Nstrm
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *Nstrm) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Nstrm) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *Nstrm) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)

		case ruleAction0:
			p.Current.FirstFilter = true
		case ruleAction1:
			p.pipeStart(begin, end)
		case ruleAction2:
			p.pipePush(begin, end)
		case ruleAction3:
			p.Current.LastFilter = true
		case ruleAction4:
			p.pipeEnd()
		case ruleAction5:
			p.addOp2("or", begin, end)
		case ruleAction6:
			p.addOp2("and", begin, end)
		case ruleAction7:
			p.addOp2("==", begin, end)
		case ruleAction8:
			p.addOp2("!=", begin, end)
		case ruleAction9:
			p.addOp2("<=", begin, end)
		case ruleAction10:
			p.addOp2(">=", begin, end)
		case ruleAction11:
			p.addOp2("<", begin, end)
		case ruleAction12:
			p.addOp2(">", begin, end)
		case ruleAction13:
			p.addOp2("ADD", begin, end)
		case ruleAction14:
			p.addOp2("SUB", begin, end)
		case ruleAction15:
			p.addOp2("MUL", begin, end)
		case ruleAction16:
			p.addOp2("DIV", begin, end)
		case ruleAction17:
			p.addOp2("MOD", begin, end)
		case ruleAction18:
			p.skip()
		case ruleAction19:
			p.pushScope()
		case ruleAction20:
			p.close()
		case ruleAction21:
			p.literal(nil, begin, end)
		case ruleAction22:
			p.literal(true, begin, end)
		case ruleAction23:
			p.literal(false, begin, end)
		case ruleAction24:
			p.prepare(buffer[begin:end])
		case ruleAction25:
			p.addArgment(buffer[begin:end])
		case ruleAction26:
			p.bind()
		case ruleAction27:
			p.refVar(buffer[begin:end], begin, end)
		case ruleAction28:
			p.funcall(begin, end)
		case ruleAction29:
			p.pushScope()
		case ruleAction30:
			p.array()
		case ruleAction31:
			p.pushScope()
		case ruleAction32:
			p.block()
		case ruleAction33:
			p.pushScope()
		case ruleAction34:
			p.ifCond()
		case ruleAction35:
			p.ifTrue()
		case ruleAction36:
			p.ifElse()
		case ruleAction37:
			p.ifElse()
		case ruleAction38:
			p.ifexpr()
		case ruleAction39:
			p.pushScope()
		case ruleAction40:
			p.whileCond()
		case ruleAction41:
			p.whileexpr()
		case ruleAction42:
			p.wait()
		case ruleAction43:
			p.pushScope()
		case ruleAction44:
			p.emit()
		case ruleAction45:
			p.addNumber(buffer[begin:end], begin, end)
		case ruleAction46:
			p.addNumber(buffer[begin:end], begin, end)
		case ruleAction47:
			s, _ := strconv.Unquote(buffer[begin:end])
			p.literal(s, begin, end)

		}
	}
	_, _, _ = buffer, begin, end
}

func (p *Nstrm) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 top <- <(body !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[rulebody]() {
					goto l0
				}
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
				depth--
				add(ruletop, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 body <- <(sp (expr period+ sp)* expr? sp)> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if !_rules[rulesp]() {
					goto l3
				}
			l5:
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					if !_rules[ruleexpr]() {
						goto l6
					}
					{
						position9 := position
						depth++
						{
							switch buffer[position] {
							case '#':
								if !_rules[rulecomment]() {
									goto l6
								}
								break
							case '\r':
								if buffer[position] != rune('\r') {
									goto l6
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l6
								}
								position++
								break
							default:
								if buffer[position] != rune(';') {
									goto l6
								}
								position++
								break
							}
						}

						depth--
						add(ruleperiod, position9)
					}
				l7:
					{
						position8, tokenIndex8, depth8 := position, tokenIndex, depth
						{
							position11 := position
							depth++
							{
								switch buffer[position] {
								case '#':
									if !_rules[rulecomment]() {
										goto l8
									}
									break
								case '\r':
									if buffer[position] != rune('\r') {
										goto l8
									}
									position++
									break
								case '\n':
									if buffer[position] != rune('\n') {
										goto l8
									}
									position++
									break
								default:
									if buffer[position] != rune(';') {
										goto l8
									}
									position++
									break
								}
							}

							depth--
							add(ruleperiod, position11)
						}
						goto l7
					l8:
						position, tokenIndex, depth = position8, tokenIndex8, depth8
					}
					if !_rules[rulesp]() {
						goto l6
					}
					goto l5
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					if !_rules[ruleexpr]() {
						goto l13
					}
					goto l14
				l13:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
				}
			l14:
				if !_rules[rulesp]() {
					goto l3
				}
				depth--
				add(rulebody, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 expr <- <e0> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				{
					position17 := position
					depth++
					{
						position18, tokenIndex18, depth18 := position, tokenIndex, depth
						{
							position20, tokenIndex20, depth20 := position, tokenIndex, depth
							if buffer[position] != rune('|') {
								goto l20
							}
							position++
							{
								add(ruleAction0, position)
							}
							goto l21
						l20:
							position, tokenIndex, depth = position20, tokenIndex20, depth20
						}
					l21:
						if !_rules[rulews]() {
							goto l19
						}
						if !_rules[rulee01]() {
							goto l19
						}
						{
							add(ruleAction1, position)
						}
						if buffer[position] != rune('|') {
							goto l19
						}
						position++
						if !_rules[rulews]() {
							goto l19
						}
						if !_rules[rulee01]() {
							goto l19
						}
						{
							add(ruleAction2, position)
						}
					l24:
						{
							position25, tokenIndex25, depth25 := position, tokenIndex, depth
							if buffer[position] != rune('|') {
								goto l25
							}
							position++
							if !_rules[rulews]() {
								goto l25
							}
							if !_rules[rulee01]() {
								goto l25
							}
							{
								add(ruleAction2, position)
							}
							goto l24
						l25:
							position, tokenIndex, depth = position25, tokenIndex25, depth25
						}
						if !_rules[rulews]() {
							goto l19
						}
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							if buffer[position] != rune('|') {
								goto l28
							}
							position++
							{
								add(ruleAction3, position)
							}
							goto l29
						l28:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
						}
					l29:
						{
							add(ruleAction4, position)
						}
						goto l18
					l19:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[rulee01]() {
							goto l15
						}
					}
				l18:
					depth--
					add(rulee0, position17)
				}
				depth--
				add(ruleexpr, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 3 e0 <- <((('|' Action0)? ws e01 Action1 ('|' ws e01 Action2)+ ws ('|' Action3)? Action4) / e01)> */
		nil,
		/* 4 e01 <- <(e1 (('|' '|' sp e1 Action5) / ('&' '&' sp e1 Action6))*)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				if !_rules[rulee1]() {
					goto l33
				}
			l35:
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					{
						position37, tokenIndex37, depth37 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l38
						}
						position++
						if buffer[position] != rune('|') {
							goto l38
						}
						position++
						if !_rules[rulesp]() {
							goto l38
						}
						if !_rules[rulee1]() {
							goto l38
						}
						{
							add(ruleAction5, position)
						}
						goto l37
					l38:
						position, tokenIndex, depth = position37, tokenIndex37, depth37
						if buffer[position] != rune('&') {
							goto l36
						}
						position++
						if buffer[position] != rune('&') {
							goto l36
						}
						position++
						if !_rules[rulesp]() {
							goto l36
						}
						if !_rules[rulee1]() {
							goto l36
						}
						{
							add(ruleAction6, position)
						}
					}
				l37:
					goto l35
				l36:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
				}
				depth--
				add(rulee01, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 5 e1 <- <(e2 (('<' '=' sp e2 Action9) / ('>' '=' sp e2 Action10) / ((&('>') ('>' sp e2 Action12)) | (&('<') ('<' sp e2 Action11)) | (&('!') ('!' '=' sp e2 Action8)) | (&('=') ('=' '=' sp e2 Action7))))*)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				if !_rules[rulee2]() {
					goto l41
				}
			l43:
				{
					position44, tokenIndex44, depth44 := position, tokenIndex, depth
					{
						position45, tokenIndex45, depth45 := position, tokenIndex, depth
						if buffer[position] != rune('<') {
							goto l46
						}
						position++
						if buffer[position] != rune('=') {
							goto l46
						}
						position++
						if !_rules[rulesp]() {
							goto l46
						}
						if !_rules[rulee2]() {
							goto l46
						}
						{
							add(ruleAction9, position)
						}
						goto l45
					l46:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
						if buffer[position] != rune('>') {
							goto l48
						}
						position++
						if buffer[position] != rune('=') {
							goto l48
						}
						position++
						if !_rules[rulesp]() {
							goto l48
						}
						if !_rules[rulee2]() {
							goto l48
						}
						{
							add(ruleAction10, position)
						}
						goto l45
					l48:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
						{
							switch buffer[position] {
							case '>':
								if buffer[position] != rune('>') {
									goto l44
								}
								position++
								if !_rules[rulesp]() {
									goto l44
								}
								if !_rules[rulee2]() {
									goto l44
								}
								{
									add(ruleAction12, position)
								}
								break
							case '<':
								if buffer[position] != rune('<') {
									goto l44
								}
								position++
								if !_rules[rulesp]() {
									goto l44
								}
								if !_rules[rulee2]() {
									goto l44
								}
								{
									add(ruleAction11, position)
								}
								break
							case '!':
								if buffer[position] != rune('!') {
									goto l44
								}
								position++
								if buffer[position] != rune('=') {
									goto l44
								}
								position++
								if !_rules[rulesp]() {
									goto l44
								}
								if !_rules[rulee2]() {
									goto l44
								}
								{
									add(ruleAction8, position)
								}
								break
							default:
								if buffer[position] != rune('=') {
									goto l44
								}
								position++
								if buffer[position] != rune('=') {
									goto l44
								}
								position++
								if !_rules[rulesp]() {
									goto l44
								}
								if !_rules[rulee2]() {
									goto l44
								}
								{
									add(ruleAction7, position)
								}
								break
							}
						}

					}
				l45:
					goto l43
				l44:
					position, tokenIndex, depth = position44, tokenIndex44, depth44
				}
				depth--
				add(rulee1, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 6 e2 <- <(e3 (('+' sp e3 Action13) / ('-' sp e3 Action14))*)> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				if !_rules[rulee3]() {
					goto l55
				}
			l57:
				{
					position58, tokenIndex58, depth58 := position, tokenIndex, depth
					{
						position59, tokenIndex59, depth59 := position, tokenIndex, depth
						if buffer[position] != rune('+') {
							goto l60
						}
						position++
						if !_rules[rulesp]() {
							goto l60
						}
						if !_rules[rulee3]() {
							goto l60
						}
						{
							add(ruleAction13, position)
						}
						goto l59
					l60:
						position, tokenIndex, depth = position59, tokenIndex59, depth59
						if buffer[position] != rune('-') {
							goto l58
						}
						position++
						if !_rules[rulesp]() {
							goto l58
						}
						if !_rules[rulee3]() {
							goto l58
						}
						{
							add(ruleAction14, position)
						}
					}
				l59:
					goto l57
				l58:
					position, tokenIndex, depth = position58, tokenIndex58, depth58
				}
				depth--
				add(rulee2, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 7 e3 <- <(e4 ((&('%') ('%' sp e4 Action17)) | (&('/') ('/' sp e4 Action16)) | (&('*') ('*' sp e4 Action15)))*)> */
		func() bool {
			position63, tokenIndex63, depth63 := position, tokenIndex, depth
			{
				position64 := position
				depth++
				if !_rules[rulee4]() {
					goto l63
				}
			l65:
				{
					position66, tokenIndex66, depth66 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '%':
							if buffer[position] != rune('%') {
								goto l66
							}
							position++
							if !_rules[rulesp]() {
								goto l66
							}
							if !_rules[rulee4]() {
								goto l66
							}
							{
								add(ruleAction17, position)
							}
							break
						case '/':
							if buffer[position] != rune('/') {
								goto l66
							}
							position++
							if !_rules[rulesp]() {
								goto l66
							}
							if !_rules[rulee4]() {
								goto l66
							}
							{
								add(ruleAction16, position)
							}
							break
						default:
							if buffer[position] != rune('*') {
								goto l66
							}
							position++
							if !_rules[rulesp]() {
								goto l66
							}
							if !_rules[rulee4]() {
								goto l66
							}
							{
								add(ruleAction15, position)
							}
							break
						}
					}

					goto l65
				l66:
					position, tokenIndex, depth = position66, tokenIndex66, depth66
				}
				depth--
				add(rulee3, position64)
			}
			return true
		l63:
			position, tokenIndex, depth = position63, tokenIndex63, depth63
			return false
		},
		/* 8 e4 <- <(value ((&('#') comment) | (&('\t') '\t') | (&(' ') ' '))*)> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				{
					position73 := position
					depth++
					{
						position74, tokenIndex74, depth74 := position, tokenIndex, depth
						{
							position76 := position
							depth++
							{
								position77 := position
								depth++
								{
									position78, tokenIndex78, depth78 := position, tokenIndex, depth
									if !_rules[ruleminus]() {
										goto l78
									}
									goto l79
								l78:
									position, tokenIndex, depth = position78, tokenIndex78, depth78
								}
							l79:
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l75
								}
								position++
							l80:
								{
									position81, tokenIndex81, depth81 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l81
									}
									position++
									goto l80
								l81:
									position, tokenIndex, depth = position81, tokenIndex81, depth81
								}
								if buffer[position] != rune('.') {
									goto l75
								}
								position++
							l82:
								{
									position83, tokenIndex83, depth83 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l83
									}
									position++
									goto l82
								l83:
									position, tokenIndex, depth = position83, tokenIndex83, depth83
								}
								depth--
								add(rulePegText, position77)
							}
							{
								add(ruleAction45, position)
							}
							depth--
							add(rulefloating, position76)
						}
						goto l74
					l75:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						if !_rules[ruleifexpr]() {
							goto l85
						}
						goto l74
					l85:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position87 := position
							depth++
							if buffer[position] != rune('w') {
								goto l86
							}
							position++
							if buffer[position] != rune('h') {
								goto l86
							}
							position++
							if buffer[position] != rune('i') {
								goto l86
							}
							position++
							if buffer[position] != rune('l') {
								goto l86
							}
							position++
							if buffer[position] != rune('e') {
								goto l86
							}
							position++
							{
								add(ruleAction39, position)
							}
							if !_rules[rulesp]() {
								goto l86
							}
							if !_rules[ruleexpr]() {
								goto l86
							}
							{
								add(ruleAction40, position)
							}
							if buffer[position] != rune('{') {
								goto l86
							}
							position++
							if !_rules[rulebody]() {
								goto l86
							}
							{
								add(ruleAction41, position)
							}
							if buffer[position] != rune('}') {
								goto l86
							}
							position++
							depth--
							add(rulewhileexpr, position87)
						}
						goto l74
					l86:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position92 := position
							depth++
							if buffer[position] != rune('e') {
								goto l91
							}
							position++
							if buffer[position] != rune('m') {
								goto l91
							}
							position++
							if buffer[position] != rune('i') {
								goto l91
							}
							position++
							if buffer[position] != rune('t') {
								goto l91
							}
							position++
							{
								add(ruleAction43, position)
							}
							if !_rules[rulesp]() {
								goto l91
							}
							{
								position94, tokenIndex94, depth94 := position, tokenIndex, depth
								if !_rules[ruleexpr]() {
									goto l95
								}
								if buffer[position] != rune(',') {
									goto l95
								}
								position++
								if !_rules[rulesp]() {
									goto l95
								}
							l96:
								{
									position97, tokenIndex97, depth97 := position, tokenIndex, depth
									if !_rules[ruleexpr]() {
										goto l97
									}
									if buffer[position] != rune(',') {
										goto l97
									}
									position++
									if !_rules[rulesp]() {
										goto l97
									}
									goto l96
								l97:
									position, tokenIndex, depth = position97, tokenIndex97, depth97
								}
								if !_rules[ruleexpr]() {
									goto l95
								}
								goto l94
							l95:
								position, tokenIndex, depth = position94, tokenIndex94, depth94
								if !_rules[ruleexpr]() {
									goto l91
								}
							}
						l94:
							{
								add(ruleAction44, position)
							}
							depth--
							add(ruleemit, position92)
						}
						goto l74
					l91:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						if buffer[position] != rune('s') {
							goto l99
						}
						position++
						if buffer[position] != rune('k') {
							goto l99
						}
						position++
						if buffer[position] != rune('i') {
							goto l99
						}
						position++
						if buffer[position] != rune('p') {
							goto l99
						}
						position++
						{
							add(ruleAction18, position)
						}
						goto l74
					l99:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						if buffer[position] != rune('c') {
							goto l101
						}
						position++
						if buffer[position] != rune('l') {
							goto l101
						}
						position++
						if buffer[position] != rune('o') {
							goto l101
						}
						position++
						if buffer[position] != rune('s') {
							goto l101
						}
						position++
						if buffer[position] != rune('e') {
							goto l101
						}
						position++
						{
							add(ruleAction19, position)
						}
						if !_rules[rulews]() {
							goto l101
						}
						{
							position103, tokenIndex103, depth103 := position, tokenIndex, depth
							if !_rules[ruleexpr]() {
								goto l103
							}
							goto l104
						l103:
							position, tokenIndex, depth = position103, tokenIndex103, depth103
						}
					l104:
						{
							add(ruleAction20, position)
						}
						goto l74
					l101:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position107 := position
							depth++
							if buffer[position] != rune('n') {
								goto l106
							}
							position++
							if buffer[position] != rune('i') {
								goto l106
							}
							position++
							if buffer[position] != rune('l') {
								goto l106
							}
							position++
							depth--
							add(rulePegText, position107)
						}
						{
							add(ruleAction21, position)
						}
						goto l74
					l106:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position110 := position
							depth++
							if buffer[position] != rune('t') {
								goto l109
							}
							position++
							if buffer[position] != rune('r') {
								goto l109
							}
							position++
							if buffer[position] != rune('u') {
								goto l109
							}
							position++
							if buffer[position] != rune('e') {
								goto l109
							}
							position++
							depth--
							add(rulePegText, position110)
						}
						{
							add(ruleAction22, position)
						}
						goto l74
					l109:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position113 := position
							depth++
							if buffer[position] != rune('f') {
								goto l112
							}
							position++
							if buffer[position] != rune('a') {
								goto l112
							}
							position++
							if buffer[position] != rune('l') {
								goto l112
							}
							position++
							if buffer[position] != rune('s') {
								goto l112
							}
							position++
							if buffer[position] != rune('e') {
								goto l112
							}
							position++
							depth--
							add(rulePegText, position113)
						}
						{
							add(ruleAction23, position)
						}
						goto l74
					l112:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position116 := position
							depth++
							if buffer[position] != rune('w') {
								goto l115
							}
							position++
							if buffer[position] != rune('a') {
								goto l115
							}
							position++
							if buffer[position] != rune('i') {
								goto l115
							}
							position++
							if buffer[position] != rune('t') {
								goto l115
							}
							position++
							{
								add(ruleAction42, position)
							}
							depth--
							add(rulewait, position116)
						}
						goto l74
					l115:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position119 := position
							depth++
							{
								position120 := position
								depth++
								if !_rules[ruleidentifer_prepare]() {
									goto l118
								}
								if buffer[position] != rune('(') {
									goto l118
								}
								position++
								if !_rules[rulesp]() {
									goto l118
								}
							l121:
								{
									position122, tokenIndex122, depth122 := position, tokenIndex, depth
									if !_rules[ruleexpr]() {
										goto l122
									}
									if buffer[position] != rune(',') {
										goto l122
									}
									position++
									goto l121
								l122:
									position, tokenIndex, depth = position122, tokenIndex122, depth122
								}
								{
									position123, tokenIndex123, depth123 := position, tokenIndex, depth
									if !_rules[ruleexpr]() {
										goto l123
									}
									goto l124
								l123:
									position, tokenIndex, depth = position123, tokenIndex123, depth123
								}
							l124:
								if buffer[position] != rune(')') {
									goto l118
								}
								position++
								depth--
								add(rulePegText, position120)
							}
							{
								add(ruleAction28, position)
							}
							depth--
							add(rulefuncall, position119)
						}
						goto l74
					l118:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							position127 := position
							depth++
							if !_rules[ruleidentifer_prepare]() {
								goto l126
							}
							if buffer[position] != rune('=') {
								goto l126
							}
							position++
							if !_rules[rulesp]() {
								goto l126
							}
							if !_rules[ruleexpr]() {
								goto l126
							}
							{
								add(ruleAction26, position)
							}
							depth--
							add(rulebind, position127)
						}
						goto l74
					l126:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						{
							switch buffer[position] {
							case '(':
								if buffer[position] != rune('(') {
									goto l71
								}
								position++
								if !_rules[rulesp]() {
									goto l71
								}
								if !_rules[ruleexpr]() {
									goto l71
								}
								if !_rules[rulesp]() {
									goto l71
								}
								if buffer[position] != rune(')') {
									goto l71
								}
								position++
								break
							case '{':
								{
									position130 := position
									depth++
									if buffer[position] != rune('{') {
										goto l71
									}
									position++
									{
										add(ruleAction31, position)
									}
									if !_rules[rulesp]() {
										goto l71
									}
								l132:
									{
										position133, tokenIndex133, depth133 := position, tokenIndex, depth
										if !_rules[rulesp]() {
											goto l133
										}
										if !_rules[ruleidentifer_argment]() {
											goto l133
										}
										if !_rules[rulesp]() {
											goto l133
										}
										if buffer[position] != rune(',') {
											goto l133
										}
										position++
										goto l132
									l133:
										position, tokenIndex, depth = position133, tokenIndex133, depth133
									}
									{
										position134, tokenIndex134, depth134 := position, tokenIndex, depth
										if !_rules[ruleidentifer_argment]() {
											goto l134
										}
										goto l135
									l134:
										position, tokenIndex, depth = position134, tokenIndex134, depth134
									}
								l135:
									if !_rules[rulesp]() {
										goto l71
									}
									if buffer[position] != rune('-') {
										goto l71
									}
									position++
									if buffer[position] != rune('>') {
										goto l71
									}
									position++
									if !_rules[rulebody]() {
										goto l71
									}
									if buffer[position] != rune('}') {
										goto l71
									}
									position++
									{
										add(ruleAction32, position)
									}
									depth--
									add(ruleblock, position130)
								}
								break
							case '[':
								{
									position137 := position
									depth++
									if buffer[position] != rune('[') {
										goto l71
									}
									position++
									{
										add(ruleAction29, position)
									}
									if !_rules[rulesp]() {
										goto l71
									}
								l139:
									{
										position140, tokenIndex140, depth140 := position, tokenIndex, depth
										if !_rules[rulesp]() {
											goto l140
										}
										if !_rules[ruleexpr]() {
											goto l140
										}
										if !_rules[rulesp]() {
											goto l140
										}
										if buffer[position] != rune(',') {
											goto l140
										}
										position++
										goto l139
									l140:
										position, tokenIndex, depth = position140, tokenIndex140, depth140
									}
									{
										position141, tokenIndex141, depth141 := position, tokenIndex, depth
										if !_rules[ruleexpr]() {
											goto l141
										}
										goto l142
									l141:
										position, tokenIndex, depth = position141, tokenIndex141, depth141
									}
								l142:
									if !_rules[rulesp]() {
										goto l71
									}
									if buffer[position] != rune(']') {
										goto l71
									}
									position++
									{
										add(ruleAction30, position)
									}
									depth--
									add(rulearray, position137)
								}
								break
							case '"':
								{
									position144 := position
									depth++
									{
										position145 := position
										depth++
										if buffer[position] != rune('"') {
											goto l71
										}
										position++
									l146:
										{
											position147, tokenIndex147, depth147 := position, tokenIndex, depth
											{
												position148, tokenIndex148, depth148 := position, tokenIndex, depth
												if buffer[position] != rune('"') {
													goto l148
												}
												position++
												goto l147
											l148:
												position, tokenIndex, depth = position148, tokenIndex148, depth148
											}
											if !matchDot() {
												goto l147
											}
											goto l146
										l147:
											position, tokenIndex, depth = position147, tokenIndex147, depth147
										}
										if buffer[position] != rune('"') {
											goto l71
										}
										position++
										depth--
										add(rulePegText, position145)
									}
									{
										add(ruleAction47, position)
									}
									depth--
									add(rulestringliteral, position144)
								}
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								{
									position150 := position
									depth++
									{
										position151 := position
										depth++
										{
											position152, tokenIndex152, depth152 := position, tokenIndex, depth
											if !_rules[ruleminus]() {
												goto l152
											}
											goto l153
										l152:
											position, tokenIndex, depth = position152, tokenIndex152, depth152
										}
									l153:
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l71
										}
										position++
									l154:
										{
											position155, tokenIndex155, depth155 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l155
											}
											position++
											goto l154
										l155:
											position, tokenIndex, depth = position155, tokenIndex155, depth155
										}
										depth--
										add(rulePegText, position151)
									}
									{
										add(ruleAction46, position)
									}
									depth--
									add(ruleinteger, position150)
								}
								break
							default:
								{
									position157 := position
									depth++
									{
										position158 := position
										depth++
										if !_rules[ruleidentifer]() {
											goto l71
										}
										depth--
										add(rulePegText, position158)
									}
									{
										add(ruleAction27, position)
									}
									depth--
									add(rulerefvariable, position157)
								}
								break
							}
						}

					}
				l74:
					depth--
					add(rulevalue, position73)
				}
			l160:
				{
					position161, tokenIndex161, depth161 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '#':
							if !_rules[rulecomment]() {
								goto l161
							}
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l161
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l161
							}
							position++
							break
						}
					}

					goto l160
				l161:
					position, tokenIndex, depth = position161, tokenIndex161, depth161
				}
				depth--
				add(rulee4, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 9 value <- <(floating / ifexpr / whileexpr / emit / ('s' 'k' 'i' 'p' Action18) / ('c' 'l' 'o' 's' 'e' Action19 ws expr? Action20) / (<('n' 'i' 'l')> Action21) / (<('t' 'r' 'u' 'e')> Action22) / (<('f' 'a' 'l' 's' 'e')> Action23) / wait / funcall / bind / ((&('(') ('(' sp expr sp ')')) | (&('{') block) | (&('[') array) | (&('"') stringliteral) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') integer) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') refvariable)))> */
		nil,
		/* 10 identifer <- <(((&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('_') '_') | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('_') '_') | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				{
					switch buffer[position] {
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l164
						}
						position++
						break
					case '_':
						if buffer[position] != rune('_') {
							goto l164
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l164
						}
						position++
						break
					}
				}

			l167:
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l168
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l168
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l168
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l168
							}
							position++
							break
						}
					}

					goto l167
				l168:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
				}
				depth--
				add(ruleidentifer, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 11 identifer_prepare <- <(<identifer> sp Action24)> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				{
					position172 := position
					depth++
					if !_rules[ruleidentifer]() {
						goto l170
					}
					depth--
					add(rulePegText, position172)
				}
				if !_rules[rulesp]() {
					goto l170
				}
				{
					add(ruleAction24, position)
				}
				depth--
				add(ruleidentifer_prepare, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 12 identifer_argment <- <(<identifer> sp Action25)> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				{
					position176 := position
					depth++
					if !_rules[ruleidentifer]() {
						goto l174
					}
					depth--
					add(rulePegText, position176)
				}
				if !_rules[rulesp]() {
					goto l174
				}
				{
					add(ruleAction25, position)
				}
				depth--
				add(ruleidentifer_argment, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 13 bind <- <(identifer_prepare '=' sp expr Action26)> */
		nil,
		/* 14 refvariable <- <(<identifer> Action27)> */
		nil,
		/* 15 funcall <- <(<(identifer_prepare '(' sp (expr ',')* expr? ')')> Action28)> */
		nil,
		/* 16 array <- <('[' Action29 sp (sp expr sp ',')* expr? sp ']' Action30)> */
		nil,
		/* 17 block <- <('{' Action31 sp (sp identifer_argment sp ',')* identifer_argment? sp ('-' '>') body '}' Action32)> */
		nil,
		/* 18 ifexpr <- <('i' 'f' Action33 sp expr Action34 '{' body Action35 '}' (sp ('e' 'l' 's' 'e') sp (('{' body Action36 '}') / (sp ifexpr Action37)))? Action38)> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('i') {
					goto l183
				}
				position++
				if buffer[position] != rune('f') {
					goto l183
				}
				position++
				{
					add(ruleAction33, position)
				}
				if !_rules[rulesp]() {
					goto l183
				}
				if !_rules[ruleexpr]() {
					goto l183
				}
				{
					add(ruleAction34, position)
				}
				if buffer[position] != rune('{') {
					goto l183
				}
				position++
				if !_rules[rulebody]() {
					goto l183
				}
				{
					add(ruleAction35, position)
				}
				if buffer[position] != rune('}') {
					goto l183
				}
				position++
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					if !_rules[rulesp]() {
						goto l188
					}
					if buffer[position] != rune('e') {
						goto l188
					}
					position++
					if buffer[position] != rune('l') {
						goto l188
					}
					position++
					if buffer[position] != rune('s') {
						goto l188
					}
					position++
					if buffer[position] != rune('e') {
						goto l188
					}
					position++
					if !_rules[rulesp]() {
						goto l188
					}
					{
						position190, tokenIndex190, depth190 := position, tokenIndex, depth
						if buffer[position] != rune('{') {
							goto l191
						}
						position++
						if !_rules[rulebody]() {
							goto l191
						}
						{
							add(ruleAction36, position)
						}
						if buffer[position] != rune('}') {
							goto l191
						}
						position++
						goto l190
					l191:
						position, tokenIndex, depth = position190, tokenIndex190, depth190
						if !_rules[rulesp]() {
							goto l188
						}
						if !_rules[ruleifexpr]() {
							goto l188
						}
						{
							add(ruleAction37, position)
						}
					}
				l190:
					goto l189
				l188:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
				}
			l189:
				{
					add(ruleAction38, position)
				}
				depth--
				add(ruleifexpr, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 19 whileexpr <- <('w' 'h' 'i' 'l' 'e' Action39 sp expr Action40 '{' body Action41 '}')> */
		nil,
		/* 20 wait <- <('w' 'a' 'i' 't' Action42)> */
		nil,
		/* 21 emit <- <('e' 'm' 'i' 't' Action43 sp (((expr ',' sp)+ expr) / expr) Action44)> */
		nil,
		/* 22 floating <- <(<(minus? [0-9]+ '.' [0-9]*)> Action45)> */
		nil,
		/* 23 integer <- <(<(minus? [0-9]+)> Action46)> */
		nil,
		/* 24 stringliteral <- <(<('"' (!'"' .)* '"')> Action47)> */
		nil,
		/* 25 sp <- <((&('#') comment) | (&('\r') '\r') | (&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position202 := position
				depth++
			l203:
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '#':
							if !_rules[rulecomment]() {
								goto l204
							}
							break
						case '\r':
							if buffer[position] != rune('\r') {
								goto l204
							}
							position++
							break
						case '\n':
							if buffer[position] != rune('\n') {
								goto l204
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l204
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l204
							}
							position++
							break
						}
					}

					goto l203
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
				depth--
				add(rulesp, position202)
			}
			return true
		},
		/* 26 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position207 := position
				depth++
			l208:
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					{
						position210, tokenIndex210, depth210 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l211
						}
						position++
						goto l210
					l211:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if buffer[position] != rune('\t') {
							goto l209
						}
						position++
					}
				l210:
					goto l208
				l209:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
				}
				depth--
				add(rulews, position207)
			}
			return true
		},
		/* 27 comment <- <('#' (!'\n' .)* '\n'?)> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				if buffer[position] != rune('#') {
					goto l212
				}
				position++
			l214:
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					{
						position216, tokenIndex216, depth216 := position, tokenIndex, depth
						if buffer[position] != rune('\n') {
							goto l216
						}
						position++
						goto l215
					l216:
						position, tokenIndex, depth = position216, tokenIndex216, depth216
					}
					if !matchDot() {
						goto l215
					}
					goto l214
				l215:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
				}
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l217
					}
					position++
					goto l218
				l217:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
				}
			l218:
				depth--
				add(rulecomment, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 28 period <- <((&('#') comment) | (&('\r') '\r') | (&('\n') '\n') | (&(';') ';'))> */
		nil,
		/* 30 Action0 <- <{p.Current.FirstFilter=true}> */
		nil,
		/* 31 Action1 <- <{ p.pipeStart(begin,end) }> */
		nil,
		/* 32 Action2 <- <{ p.pipePush(begin,end) }> */
		nil,
		/* 33 Action3 <- <{p.Current.LastFilter=true}> */
		nil,
		/* 34 Action4 <- <{ p.pipeEnd() }> */
		nil,
		/* 35 Action5 <- <{ p.addOp2("or",begin,end)}> */
		nil,
		/* 36 Action6 <- <{ p.addOp2("and",begin,end)}> */
		nil,
		/* 37 Action7 <- <{ p.addOp2("==",begin,end) }> */
		nil,
		/* 38 Action8 <- <{ p.addOp2("!=",begin,end) }> */
		nil,
		/* 39 Action9 <- <{ p.addOp2("<=",begin,end) }> */
		nil,
		/* 40 Action10 <- <{ p.addOp2(">=",begin,end) }> */
		nil,
		/* 41 Action11 <- <{ p.addOp2("<" ,begin,end) }> */
		nil,
		/* 42 Action12 <- <{ p.addOp2(">" ,begin,end) }> */
		nil,
		/* 43 Action13 <- <{ p.addOp2("ADD",begin,end) }> */
		nil,
		/* 44 Action14 <- <{ p.addOp2("SUB",begin,end) }> */
		nil,
		/* 45 Action15 <- <{ p.addOp2("MUL",begin,end) }> */
		nil,
		/* 46 Action16 <- <{ p.addOp2("DIV",begin,end) }> */
		nil,
		/* 47 Action17 <- <{ p.addOp2("MOD",begin,end) }> */
		nil,
		/* 48 Action18 <- <{ p.skip()  }> */
		nil,
		/* 49 Action19 <- <{ p.pushScope() }> */
		nil,
		/* 50 Action20 <- <{ p.close() }> */
		nil,
		nil,
		/* 52 Action21 <- <{ p.literal(nil,begin,end) }> */
		nil,
		/* 53 Action22 <- <{ p.literal(true,begin,end) }> */
		nil,
		/* 54 Action23 <- <{ p.literal(false,begin,end) }> */
		nil,
		/* 55 Action24 <- <{ p.prepare(buffer[begin:end]) }> */
		nil,
		/* 56 Action25 <- <{ p.addArgment(buffer[begin:end]) }> */
		nil,
		/* 57 Action26 <- <{ p.bind() }> */
		nil,
		/* 58 Action27 <- <{ p.refVar(buffer[begin:end],begin,end) }> */
		nil,
		/* 59 Action28 <- <{ p.funcall(begin,end) }> */
		nil,
		/* 60 Action29 <- <{ p.pushScope() }> */
		nil,
		/* 61 Action30 <- <{ p.array() }> */
		nil,
		/* 62 Action31 <- <{ p.pushScope() }> */
		nil,
		/* 63 Action32 <- <{ p.block() }> */
		nil,
		/* 64 Action33 <- <{ p.pushScope() }> */
		nil,
		/* 65 Action34 <- <{ p.ifCond() }> */
		nil,
		/* 66 Action35 <- <{ p.ifTrue() }> */
		nil,
		/* 67 Action36 <- <{ p.ifElse() }> */
		nil,
		/* 68 Action37 <- <{ p.ifElse() }> */
		nil,
		/* 69 Action38 <- <{ p.ifexpr() }> */
		nil,
		/* 70 Action39 <- <{ p.pushScope() }> */
		nil,
		/* 71 Action40 <- <{ p.whileCond() }> */
		nil,
		/* 72 Action41 <- <{ p.whileexpr() }> */
		nil,
		/* 73 Action42 <- <{ p.wait() }> */
		nil,
		/* 74 Action43 <- <{ p.pushScope() }> */
		nil,
		/* 75 Action44 <- <{ p.emit() }> */
		nil,
		/* 76 minus <- <> */
		func() bool {
			{
				position266 := position
				depth++
				depth--
				add(ruleminus, position266)
			}
			return true
		},
		/* 77 Action45 <- <{ p.addNumber(buffer[begin:end],begin,end) }> */
		nil,
		/* 78 Action46 <- <{ p.addNumber(buffer[begin:end],begin,end) }> */
		nil,
		/* 79 Action47 <- <{ s,_:=strconv.Unquote(buffer[begin:end]);p.literal(s,begin,end) }> */
		nil,
	}
	p.rules = _rules
}
