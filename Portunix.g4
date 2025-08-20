grammar Portunix;

// Lexer rules
LONG_OPTION: '-' [a-zA-Z_][a-zA-Z0-9_]*; 

COLON: ':';

EQUAL: '=';

WORD: [a-zA-Z0-9_./]+;

STRING: '"' (~["\r\n])* '"';

WS: [ \t\r\n]+ -> skip;

// Parser rules
program: command+;

command: WORD arguments?;

parameters: WORD EQUAL (STRING | WORD); 

arguments: (parameters | WORD | LONG_OPTION | STRING)+;





