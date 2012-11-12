package com.bitgirder.validation;

import com.bitgirder.lang.Lang;

import com.bitgirder.test.Test;

import java.util.Set;
import java.util.Properties;

@Test
final
class ValidationTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    @Test( expected = IllegalStateException.class,
           expectedPattern = "Set 's' did not contain element: s3" )
    private
    void
    testValidatorRemove()
    {
        Set< String > s = Lang.newSet( "s1", "s2" );

        state.isTrue( state.remove( s, "s1", "s" ) );
        state.equalInt( 1, s.size() );
        state.isTrue( s.contains( "s2" ) );

        // now fail
        state.remove( s, "s3", "s" );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = "Properties 'p' has no value for key prop2" )
    private
    void
    testExpectProperty()
    {
        Properties p = new Properties();
        p.setProperty( "prop1", "val1" );

        state.equalString( "val1", state.getProperty( p, "prop1", "p" ) );

        state.getProperty( p, "prop2", "p" );
    }
}
