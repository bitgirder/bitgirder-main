package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

@Test
final
class ParserTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // This is important not just for its own sake but because other test
    // classes in the codebase rely on the format of the message for their own
    // assertions. Generally speaking we wouldn't to couple the notion of
    // correctness of a program to parsing an exception message, but in these
    // cases the correctness that we're asserting *is* the exception message.
    @Test
    private
    void
    testExceptionMessageWithLocation()
    {
        SourceTextLocation loc =
            SourceTextLocation.create( "test-file", 12, 23 );
 
        state.equalString(
            "test-file [12,23]: Bad stuff happened",
            new SyntaxException( "Bad stuff happened", loc ).getMessage()
        );
    }
}
