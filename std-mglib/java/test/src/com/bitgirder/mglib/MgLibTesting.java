package com.bitgirder.mglib;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.mingle.model.MingleTests;
import com.bitgirder.mingle.model.TypeDefinitionLookup;

import com.bitgirder.mingle.bind.MingleBinders;
import com.bitgirder.mingle.bind.MingleBinder;
import com.bitgirder.mingle.bind.MingleBindTests;

import com.bitgirder.test.TestRuntime;

import com.bitgirder.testing.Testing;

public
final
class MgLibTesting
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    private MgLibTesting() {}

    public
    static
    MingleBinder
    expectDefaultBinder( TestRuntime rt )
    {
        return
            Testing.expectObject(
                inputs.notNull( rt, "rt" ),
                MingleBindTests.KEY_BINDER,
                MingleBinder.class
            );
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    init( final Testing.RuntimeInitializerContext ctx )
        throws Exception
    {
        Testing.awaitTestObject(
            ctx,
            MingleTests.KEY_TYPE_DEFINITION_LOOKUP,
            TypeDefinitionLookup.class,
            new ObjectReceiver< TypeDefinitionLookup >() {
                public void receive( TypeDefinitionLookup types )
                    throws Exception
                {
                    ctx.setObject( 
                        MingleBindTests.KEY_BINDER, 
                        MingleBinders.loadDefault( types ) );

                    ctx.complete();
                }
            }
        );
    }
}
