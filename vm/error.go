package vm

import (
	"fmt"
	"log"

	"../ast"
)

func (e Error) Fatal(buffer string) {
	log.Fatalf("\n%sError: %s\n", show(buffer, e.Pos), e.Message)
}

func show(buffer string, pos ast.Position) string {
	var line = 1
	var column = 0
	for i := 0; i < pos.Begin; i++ {
		if buffer[i] == '\n' {
			line++
			column = 0
		}
		column++
	}
	return fmt.Sprintf("line: %d, Column: %d\n%s\n", line, column, buffer[pos.Begin:pos.End])
}

func Errorf(p ast.Pos, str string, args ...interface{}) *Error {
	return &Error{
		Pos:     p.GetPosition(),
		Message: fmt.Sprintf(str, args...),
	}
}
