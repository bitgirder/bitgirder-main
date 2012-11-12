package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.ModelTestInstances;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleValidationException;

import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.TestCall;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.TestFailureExpector;

public
abstract
class AbstractRoundtripTest< R extends AbstractRoundtripTest< R > >
implements TestFailureExpector,
           LabeledTestObject,
           TestCall
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private CharSequence lblPre;
    private Object jvObj;
    private String jPathRoot;
    private MingleTypeReference mgTyp;
    private MingleTypeReference assertTyp;
    private MingleValue mgVal;
    private MingleIdentifier mgPathRoot;
    private MingleBinder mb;
    private Class< ? extends Throwable > failCls;
    private CharSequence failPat;
    private CharSequence failLocation;
    private boolean fromJava = true;
    private boolean debug = false;
    private boolean useOpaque = false;

    protected AbstractRoundtripTest() {}

    public Object getInvocationTarget() { return this; }

    public
    CharSequence
    getLabel()
    {
        CharSequence tail =
            Strings.crossJoin( "=", ",",
                "jvObj", jvObj,
                "jPathRoot", jPathRoot,
                "mgTyp", mgTyp,
                "assertTyp", assertTyp,
                "mgVal", mgVal == null ? null : MingleModels.inspect( mgVal ),
                "mgPathRoot", mgPathRoot,
                "failCls", failCls,
                "failPat", failPat,
                "fromJava", fromJava,
                "useOpaque", useOpaque
            );
        
        return lblPre == null ? tail : lblPre.toString() + "," + tail;
    } 

    public
    final
    Class< ? extends Throwable >
    expectedFailureClass()
    {
        return failCls;
    }

    public 
    final 
    CharSequence 
    expectedFailurePattern() 
    { 
        if ( failLocation == null ) return failPat;
        else return "\\Q" + failLocation + ": \\E" + failPat;
    }

    private R castThis() { return Lang.< R >castUnchecked( this ); }

    public
    final
    R
    setLabelPrefix( CharSequence lblPre )
    {
        this.lblPre = inputs.notNull( lblPre, "lblPre" );
        return castThis();
    }

    public
    final
    R
    setJvObj( Object jvObj ) 
    {
        this.jvObj = jvObj;
        return castThis();
    }

    public
    final
    R
    setJvPathRoot( String root )
    {
        this.jPathRoot = root;
        return castThis();
    }

    public
    final
    R
    setMgType( MingleTypeReference mgTyp )
    {
        this.mgTyp = mgTyp;
        return castThis();
    }

    public
    final
    R
    setMgType( CharSequence mgTyp )
    {
        return setMgType( MingleTypeReference.create( mgTyp ) );
    }

    public
    final
    R
    setAssertType( MingleTypeReference assertTyp )
    {
        this.assertTyp = inputs.notNull( assertTyp, "assertTyp" );
        return castThis();
    }

    public
    final
    R
    setAssertType( CharSequence str )
    {
        inputs.notNull( str, "str" );
        return setAssertType( MingleTypeReference.create( str ) );
    }

    public
    final
    R
    setMgVal( MingleValue mgVal )
    {
        this.mgVal = mgVal;
        return castThis();
    }

    public
    final
    R
    setMgPathRoot( MingleIdentifier mgPathRoot )
    {
        this.mgPathRoot = inputs.notNull( mgPathRoot, "mgPathRoot" );
        return castThis();
    }

    public
    final
    R
    setMgPathRoot( CharSequence str )
    {
        return 
            setMgPathRoot( 
                MingleIdentifier.create( inputs.notNull( str, "str" ) ) );
    }

    public
    final
    R
    setBinder( MingleBinder mb )
    {
        this.mb = mb;
        return castThis();
    }

    public
    final
    R
    setFailureClass( Class< ? extends Throwable > cls )
    {
        this.failCls = inputs.notNull( cls, "cls" );
        return castThis();
    }

    public
    final
    R
    setFailurePattern( CharSequence failPat )
    {
        this.failPat = failPat;
        return castThis();
    }

    public
    final
    R
    setFailureLocation( CharSequence failLocation )
    {
        this.failLocation = inputs.notNull( failLocation, "failLocation" );
        return castThis();
    }

    public
    final
    R
    setFromJava() 
    { 
        fromJava = true; 
        return castThis();
    }

    public
    final
    R
    setFromMingle()
    {
        fromJava = false;
        return castThis();
    }

    public
    final
    R
    setDebug()
    {
        debug = true;
        return castThis();
    }

    public
    final
    R
    setUseOpaque()
    {
        useOpaque = true;
        return castThis();
    }

    private
    void
    validate()
    {
        state.notNull( mb, "mb" );
    }

    private
    ObjectPath< String >
    jPath()
    {
        if ( jPathRoot == null ) return ObjectPath.< String >getRoot();
        else return ObjectPath.getRoot( jPathRoot );
    }

    private
    ObjectPath< MingleIdentifier >
    mgPath()
    {
        if ( mgPathRoot == null ) 
        {
            return ObjectPath.< MingleIdentifier >getRoot();
        }
        else return ObjectPath.getRoot( mgPathRoot );
    }

    protected
    abstract
    void
    assertJavaValues( Object jvObj1,
                      Object jvObj2,
                      MingleTypeReference mgTyp )
        throws Exception;

    private
    void
    doAssertJavaValues( Object jvObj2 )
        throws Exception
    {
        assertJavaValues( 
            jvObj, jvObj2, assertTyp == null ? mgTyp : assertTyp );
    }

    private
    MingleValue
    asMingleValue( Object o )
    {
        return MingleBinders.asMingleValue( mb, mgTyp, o, jPath(), useOpaque );
    }

    private
    Object
    asJavaValue( MingleValue mv )
    {
        return MingleBinders.asJavaValue( mb, mgTyp, mv, mgPath(), useOpaque );
    }

    private
    void
    callFromJava()
        throws Exception
    {
        MingleValue mgValActual = asMingleValue( jvObj );

        if ( debug ) 
        {
            code( "mgValActual:", MingleModels.inspect( mgValActual ) );
        }

        ModelTestInstances.assertEqual( mgVal, mgValActual );

        Object jvObj2 = asJavaValue( mgValActual );

        doAssertJavaValues( jvObj2 );
    }

    private
    void
    callFromMingle()
        throws Exception
    {
        Object jvObj2 = asJavaValue( mgVal );
        doAssertJavaValues( jvObj2 );

        MingleValue mgVal2 = asMingleValue( jvObj2 );

        if ( debug ) code( "mgVal2:", MingleModels.inspect( mgVal2 ) );
        ModelTestInstances.assertEqual( mgVal, mgVal2 );
    }

    public
    final
    void
    call( TestRuntime rt )
        throws Exception
    {
        validate();

        if ( fromJava ) callFromJava(); else callFromMingle();
    }
}
