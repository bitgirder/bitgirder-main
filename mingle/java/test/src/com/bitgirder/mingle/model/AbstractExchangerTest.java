package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.test.LabeledTestCall;

public
abstract
class AbstractExchangerTest< T extends AbstractExchangerTest >
extends LabeledTestCall
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }
 
    private MingleValueExchanger exch;
    private Object jvObj;
    private MingleValue mgVal;
    
    protected AbstractExchangerTest( CharSequence lbl ) { super( lbl ); }

    protected
    abstract
    void
    assertExchange( Object jvObj1,
                    Object jvObj2 )
        throws Exception;

    private T castThis() { return Lang.< T >castUnchecked( this ); }

    public
    final
    T
    setExchanger( MingleValueExchanger exch )
    {
        this.exch = exch;
        return castThis();
    }

    public
    final
    T
    setJvObj( Object jvObj )
    {
        this.jvObj = jvObj;
        return castThis();
    }

    public
    final
    T
    setMgVal( MingleValue mgVal )
    {
        this.mgVal = mgVal;
        return castThis();
    }

    // For tests when we're only after inbound mingle exceptions, jvObj will be
    // null and mgVal will be our starting point for the (presumably to fail)
    // conversion from mingle --> java; otherwise we convert jvObj to mgVal2,
    // check its equality with mgVal, and return it
    private
    MingleValue
    getMgVal2()
        throws Exception
    {
        if ( jvObj == null ) return mgVal;
        else
        {
            ObjectPath< String > jvPath = ObjectPath.getRoot();
            MingleValue mgVal2 = exch.asMingleValue( jvObj, jvPath );
            ModelTestInstances.assertEqual( mgVal, mgVal2 );
            
            return mgVal2;
        }
    }

    public
    final
    void
    call()
        throws Exception
    {
        state.notNull( exch, "exch" );
        state.notNull( mgVal, "mgVal" );
        
        MingleValue mgVal2 = getMgVal2();

        ObjectPath< MingleIdentifier > mgPath = ObjectPath.getRoot();
        Object jvObj2 = exch.asJavaValue( mgVal2, mgPath );

        if ( state.sameNullity( jvObj, jvObj2 ) ) 
        {
            assertExchange( jvObj, jvObj2 );
        }
    }
}
