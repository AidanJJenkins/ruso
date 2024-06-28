package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"

	INTEGER = "INTEGER"
	STRING  = "STRING"
	BOOLEAN = "BOOLEAN"
	FALSE   = "FALSE"
	TRUE    = "TRUE"

	ASSIGN = "="
	ALL    = "*"

	COMMA     = ","
	SEMICOLON = ";"
	QUOTE     = "\""

	LPAREN = "("
	RPAREN = ")"

	INSERT  = "INSERT"
	INTO    = "INTO"
	SELECT  = "SELECT"
	UPDATE  = "UPDATE"
	DELETE  = "DELETE"
	CREATE  = "CREATE"
	TABLE   = "TABLE"
	INDEX   = "INDEX"
	ON      = "ON"
	SET     = "SET"
	WHERE   = "WHERE"
	FROM    = "FROM"
	VALUES  = "VALUES"
	VARCHAR = "VARCHAR"
	BOOL    = "BOOL"
	INT     = "INT"
	UNIQUE  = "UNIQUE"
	AND     = "AND"
)

var keywords = map[string]TokenType{
	"INSERT":  INSERT,
	"INTO":    INTO,
	"SELECT":  SELECT,
	"UPDATE":  UPDATE,
	"DELETE":  DELETE,
	"CREATE":  CREATE,
	"TABLE":   TABLE,
	"INDEX":   INDEX,
	"ON":      ON,
	"SET":     SET,
	"WHERE":   WHERE,
	"FROM":    FROM,
	"VALUES":  VALUES,
	"UNIQUE":  UNIQUE,
	"VARCHAR": VARCHAR,
	"BOOL":    BOOL,
	"INT":     INT,
	"AND":     AND,
}

func LookupIdentifierType(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
