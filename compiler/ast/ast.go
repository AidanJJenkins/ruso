package ast

import "github.com/aidanjjenkins/compiler/token"

type Node interface {
	TokenLiteral() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

type SelectStatement struct {
	Token     token.Token
	TName     *Identifier
	Condition []*Condition
}

type Identifier struct {
	Token token.Token
	Val   string
}

type Condition struct {
	CName  *Identifier
	CIdent *Identifier
	And    *Condition
}

func (ss *SelectStatement) statementNode()       {}
func (ss *SelectStatement) TokenLiteral() string { return ss.Token.Literal }

type DeleteStatement struct {
	Token     token.Token
	TName     *Identifier
	Condition []*Condition
}

func (ds *DeleteStatement) statementNode()       {}
func (ds *DeleteStatement) TokenLiteral() string { return ds.Token.Literal }

type InsertStatement struct {
	Token  token.Token
	TName  *Identifier
	Values []string
}

func (is *InsertStatement) statementNode()       {}
func (is *InsertStatement) TokenLiteral() string { return is.Token.Literal }

type UpdateStatement struct {
	Token     token.Token
	TName     *Identifier
	Cols      []*Identifier
	Values    []string
	Condition []*Condition
}

func (us *UpdateStatement) statementNode()       {}
func (us *UpdateStatement) TokenLiteral() string { return us.Token.Literal }

type CreateTableStatement struct {
	Token    token.Token
	TName    *Identifier
	Cols     []string
	ColTypes []string
}

func (cts *CreateTableStatement) statementNode()       {}
func (cts *CreateTableStatement) TokenLiteral() string { return cts.Token.Literal }

type CreateIndexStatement struct {
	Token token.Token
	TName *Identifier
	Cols  []*Identifier
}

func (cis *CreateIndexStatement) statementNode()       {}
func (cis *CreateIndexStatement) TokenLiteral() string { return cis.Token.Literal }
