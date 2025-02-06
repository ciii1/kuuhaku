package khk

const KHK = `##########
# TOKENS #
##########

`+"`"+``+"`"+`
	function convertToNewLines(str, tolerance)
		if tolerance == -1 then
			return str
		end
		if tolerance == -2 then
			return ""
		end
		nlCount = 0
		for i = 1, str:len() do
			if str:sub(i, i) == "\n" then
				nlCount = nlCount + 1
			end
		end
		out = ""
		if nlCount >= tolerance then
			out = "\n"
		end
		return out
	end
`+"`"+``+"`"+`

IDENTIFIER { <[_a-zA-Z]+[_a-zA-Z0-9]*> }

OPENING_CURLY_BRACKET { <{> }

CLOSING_CURLY_BRACKET { <}> }

LUA_LITERAL { 
	<`+"`"+``+"`"+`> <([^`+"`"+`\\]|\\.)*> <`+"`"+``+"`"+`> 
	= 
	`+"`"+``+"`"+`
		return LITERAL1 .. "\n\t\t" .. LITERAL2:match("^%s*(.-)%s*$") .. "\n\t" .. LITERAL3
	`+"`"+``+"`"+`
}

GLOBAL_LUA_LITERAL { 
	<`+"`"+``+"`"+`> <([^`+"`"+`\\]|\\.)*> <`+"`"+``+"`"+`> 
	= 
	`+"`"+``+"`"+`
		return LITERAL1 .. "\n\t" .. LITERAL2:match("^%s*(.-)%s*$") .. "\n" .. LITERAL3
	`+"`"+``+"`"+`
}

LUA_RETURN_LITERAL { <`+"`"+`([^`+"`"+`\\\n]|\\.)*`+"`"+`>}

REGEX_LITERAL { <\<([^\>\\]|\\.)*\>> }

LuaLiteral {
	LUA_LITERAL
}
LuaLiteral {
	LUA_RETURN_LITERAL
}

GlobalLuaLiteral {
	GLOBAL_LUA_LITERAL
}
GlobalLuaLiteral {
	LUA_RETURN_LITERAL
}

EQUAL { <=> }

OPENING_BRACKET { <\(> }

CLOSING_BRACKET { <\)> }

COMMA { <\,> }

c {<#[^\n\r]*(?=(\n|\r|$))> = `+"`"+`LITERAL1`+"`"+`}

wS { wS WS}
wS { WS }

WS { <[ \t\n\r]+> = `+"`"+`convertToNewLines(LITERAL1, 2)`+"`"+` }
WS { c = `+"`"+`c1 .. "\n"`+"`"+`}

w (tolerance) { W(`+"`"+`-1`+"`"+`) w(`+"`"+`tolerance`+"`"+`) }
w (tolerance) { W(`+"`"+`tolerance`+"`"+`) }

W (tolerance) { 
	<[ \t\n\r]*> 
	= 
	`+"`"+``+"`"+`
		return convertToNewLines(LITERAL1, tolerance)
	`+"`"+``+"`"+`
}
W (tolerance) { c }

#############
# TOP LEVEL #
#############

Start {
	wS Rules w(`+"`"+`-2`+"`"+`)
}

Start {
	Rules w(`+"`"+`-2`+"`"+`)
}

Rules {
	Rules w(`+"`"+`1`+"`"+`) Rule = `+"`"+`Rules1 .. w1 .. "\n" .. Rule1`+"`"+`
}

Rules {
	Rule
}

Rules {
	GlobalLuaLiteral
}

Rule {
	Lhs OPENING_CURLY_BRACKET w(`+"`"+`-2`+"`"+`) RuleBody CLOSING_CURLY_BRACKET 
	= 
	`+"`"+`Lhs1 .. " " .. OPENING_CURLY_BRACKET1 .. w1 .. "\n" .. RuleBody1 .. "\n" .. CLOSING_CURLY_BRACKET1`+"`"+`
}

RuleBody {
	MatchRules
	= 
	`+"`"+``+"`"+`
		out = MatchRules1
		if MatchRules1:sub(MatchRules1:len(), MatchRules1:len()) == "\n" then
			out = MatchRules1:sub(1, MatchRules1:len()-1)	
		end
		return "\t" .. out
	`+"`"+``+"`"+`
}

RuleBody {
	MatchRules EQUAL w(`+"`"+`-2`+"`"+`) LuaLiteral w(`+"`"+`-2`+"`"+`)
	=
	`+"`"+``+"`"+`
		out = MatchRules1
		if MatchRules1:sub(MatchRules1:len(), MatchRules1:len()) == "\n" then
			out = MatchRules1:sub(1, MatchRules1:len()-1)	
		end
		return "\t" .. out .. "\n\t" .. EQUAL1 .. w1 .. "\n\t" .. LuaLiteral1 .. w2
	`+"`"+``+"`"+`
}

###################
# LHS AND PARAMS #
###################

Lhs {
	IDENTIFIER w(`+"`"+`-2`+"`"+`)
}

Lhs {
	IDENTIFIER w(`+"`"+`-2`+"`"+`) OPENING_BRACKET w(`+"`"+`2`+"`"+`) CommaSeparatedIdentifiers w(`+"`"+`2`+"`"+`) CLOSING_BRACKET w(`+"`"+`-2`+"`"+`)
}

CommaSeparatedIdentifiers {
	IDENTIFIER
}

CommaSeparatedIdentifiers {
	CommaSeparatedIdentifiers w(`+"`"+`-2`+"`"+`) COMMA w(`+"`"+`1`+"`"+`) IDENTIFIER
}

###############
# MATCH RULES #
###############

MatchRules {
	MatchRules MatchRule
	=
	`+"`"+`MatchRules1 .. " " .. MatchRule1`+"`"+`
}

MatchRules {
	MatchRule
}

MatchRule {
	Identifier
}

MatchRule {
	REGEX_LITERAL w(`+"`"+`1`+"`"+`)
}

########################
# IDENTIFIER WITH ARGS #
########################

Identifier {
	IDENTIFIER w(`+"`"+`-2`+"`"+`) OPENING_BRACKET w(`+"`"+`1`+"`"+`) CommaSeparatedLua w(`+"`"+`1`+"`"+`) CLOSING_BRACKET w(`+"`"+`1`+"`"+`)
}

Identifier {
	IDENTIFIER w(`+"`"+`1`+"`"+`)
}

CommaSeparatedLua {
	LuaLiteral
}

CommaSeparatedLua {
	CommaSeparatedLua w(`+"`"+`-2`+"`"+`) COMMA w(`+"`"+`1`+"`"+`) LuaLiteral
}`
