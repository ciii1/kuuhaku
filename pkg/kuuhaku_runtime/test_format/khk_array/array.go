package khk_array

const ARRAY =` 
IDENTIFIER { <[a-zA-Z]+> }

OPENING_CURLY_BRACKET { <{> }

CLOSING_CURLY_BRACKET { <}> }

w { <[ \t\n\r]*> }
wS { <[\t\n\r]*> }

Start {wS Arrays w = `+"`"+`Arrays1`+"`"+`}
Start {Arrays}

Arrays {
	Arrays w Array = `+"`"+``+"`"+`return Arrays1 .. "\n" .. Array1`+"`"+``+"`"+`
}

Arrays {
	Array
}

Array {
	OPENING_CURLY_BRACKET w Elements w CLOSING_CURLY_BRACKET 
	= `+"`"+``+"`"+`
		return OPENING_CURLY_BRACKET1 .. "\n" .. Elements1 .. "\n" .. CLOSING_CURLY_BRACKET1
	`+"`"+``+"`"+`
}

Elements {
	IDENTIFIER = `+"`"+`"\t" .. IDENTIFIER1`+"`"+`
}

Elements {
	Elements w IDENTIFIER
	= `+"`"+``+"`"+`
		return Elements1 .. "\n\t" .. IDENTIFIER1
	`+"`"+``+"`"+`
}`
