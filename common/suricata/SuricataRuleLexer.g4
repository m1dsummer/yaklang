lexer grammar SuricataRuleLexer;

// Keywords
Any: 'any';

// Symbols
Negative: '!';
Dollar: '$';
Arrow: '->';
Mul: '*';
Div: '/';
Mod: '%';
Amp: '&';
Plus: '+';
Sub: '-';
Power: '^';
Lt: '<';
Gt: '>';
LtEq: '<=';
GtEq: '>=';
Colon: ':';
DoubleColon: '::';
LBracket: '[';
RBracket: ']';
ParamStart: '(' -> pushMode(PARAM_MODE);
LBrace: '{';
RBrace: '}';
Comma: ',';
Eq: '=';
NotSymbol: '~';
Dot: '.';

LINE_COMMENT: ('#' | '//') SingleLineInputCharacter* -> skip;

ID
    : [a-zA-Z_][a-zA-Z_0-9]*
    ;

NORMALSTRING
    : '"' ( EscapeSequence | ~('\\'|'"') )* '"'
    ;

INT
    : Digit+
    ;

HEX
    : HexDigit+
    ;

FLOAT
    : Digit+ '.' Digit* ExponentPart?
    | '.' Digit+ ExponentPart?
    | Digit+ ExponentPart
    ;

fragment
ExponentPart
    : [eE] [+-]? Digit+
    ;

fragment
HexExponentPart
    : [pP] [+-]? Digit+
    ;

fragment
EscapeSequence
    : '\\' [abfnrtvz"'|$#\\]   // World of Warcraft Lua additionally escapes |$#
    | '\\' '\r'? '\n'
    | DecimalEscape
    | HexEscape
    | UtfEscape
    ;

fragment
DecimalEscape
    : '\\' Digit
    | '\\' Digit Digit
    | '\\' [0-2] Digit Digit
    ;

fragment
HexEscape
    : '\\' 'x' HexDigit HexDigit
    ;

fragment
UtfEscape
    : '\\' 'u{' HexDigit+ '}'
    ;

fragment
Digit
    : [0-9]
    ;

fragment
HexDigit
    : [0-9a-fA-F]
    ;

fragment
SingleLineInputCharacter
    : ~[\r\n\u0085\u2028\u2029]
    ;

WS
    : [ \t\u000C\r\n]+ -> skip
    ;

NonSemiColon: [^;]+;

SHEBANG
    : '#' '!' SingleLineInputCharacter* -> channel(HIDDEN)
    ;

mode PARAM_MODE;
    fragment Quote: '"';
    fragment CharInQuotedString: ~["] ;
    ParamQuotedString: Quote CharInQuotedString* Quote;
    ParamSep: ';';
    ParamValue: FreeValueAnyChar+;
    ParamEnd: ')' -> popMode;
    fragment FreeValueAnyChar: ~(')' | '"' | ';');
