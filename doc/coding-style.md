# BitGirder Coding Style Guide

Following is a set of guidelines for coding in the BitGirder codebase. This
document provides some general editing guidelines, and also gives
language-specific advice. 

While many languages have their own conventions and idioms, the BitGirder style
guidelines, like the codebase itself, takes as its fundamental goal to provide
uniformity across languages. For that reason, in most cases when given a choice
between our own style and that which is the accepted norm in a specific
language, we choose our own. In cases when the language itself requires a
certain style or syntax, then of course we have no choice.

Note that these guidelines are fluid and moving as we evolve, so in situations
where existing code does not fit the guidelines, it is okay to leave it if you
are not working with it. Otherwise: refactor it as you go to conform to the
correct (according to this document) style.

## General Coding and Editing

Unless otherwise stated, the following apply in all languages:

-   Instead of tabs, use 4 spaces

-   Wrap lines at 80 columns

-   Assume a terminal height of 24 lines, and do not let a function body
    (declaration not included) be longer than this. If a function outgrows 24
    lines, it probably needs to be broken up into smaller functions. Isolated
    exceptions include `switch` or `if/else` blocks with many possible match
    options. In these cases, the match bodies themselves should be compact (1-2
    lines).

-   Use whitespace around all operators and separators:
    
        // Bad
        a=b+c[1]
        d=f(a)
        if (d==2) {stuff()}
    
        // Good
        a = b + c[ 1 ]
        d = f( a )
        if ( d == 2 ) { stuff() }

-   Wrap long statements to avoid spanning 80 columns, indenting the remainder
    of the statement according to the type of statement:

        aVal = aLongThing( a, b, c + 3 ) + anotherLongThing( xy ) -
            someMoreLongThings() + "a string";

        if ( aLongThing( a ) == 2 || anotherLongThing( 2 ) > 3 ||
             someMoreLongThings() != k )
        {
            doSomeStuff()
        }

        for ( i = someLongThing(); aLongConditionCheck() && anotherCheck();
              i++ )
        {
            doSomeStuff()
        }

-   Put `{` and `}` on their own line when they contain blocks that are anything
    other than a sequence of single line statements, or when the block is part
    of a statement which has an introduction spanning more than one line:
        
        // 'if' is more than one line
        if ( aLongThing( a ) == 2 || anotherLongThing( 2 ) > 3 ||
             someMoreLongThings() != k )
        {
            doSomeStuff()
        }

        // 'for' is more than one line
        for ( i = someLongThing(); aLongConditionCheck() && anotherCheck();
              i++ )
        {
            doSomeStuff()
        }

        // okay block is a series of single-line statements
        if ( aLongThing( a ) == 2 || anotherLongThing( 2 ) > 3 ) {
            doAThing( a, b, c + d*2, "a string argument" );
        }

        // okay block is a series of single-line statements
        for ( a = 0; a < 100; ++a ) {
            something( a );
            somethingElse();
        }

        // block has empty lines and comments -- needs more spacing
        for ( a = 0; a < 100; ++a ) 
        {
            // we call this for each value
            something( a );

            somethingElse();
        }

   In languages without braces (such as Ruby) or where the brace needs to be on
   the same line as the statement (Go, JavaScript), leave a blank line when a
   brace would otherwise be on its own line:

        if ( aLongThing( a ) == 2 || anotherLongThing( 2 ) > 3 ||
             someMoreLongThings() != k ) {
        
            doSomeStuff()
        }

-   Blocks that are a single non-block statement and can fit in 80 columns may
    drop the braces:

        if ( a ) doStuff()

        for ( i = 0; i < 100; ++i ) something( i );

        // statement itself contains a block -- single line not okay
        if ( a ) {
            for ( i = 0; i < 100; ++i ) something( i );
        }

-   Under no circumstances is it okay to place an un-braced block below its
    declaring condition

        // don't do this
        if ( a )
            something( i )

        // or this
        for ( i = 0; i < 100; ++i )
            something( i )
