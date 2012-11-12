package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.LangTests;

import com.bitgirder.test.Test;

import java.io.InputStream;
import java.io.ByteArrayOutputStream;

import java.lang.reflect.Method;
import java.lang.reflect.Field;
import java.lang.reflect.Constructor;

import java.util.Map;

@Test
final
class ReflectionTests
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public static void code( Object... msg ) { LangTests.code( msg ); }

    private final static Object MARKER_VAL_KEY = new Object();
    private final static Object MARKER_VAL1 = "marker1-val";
    private final static Object RETURN_VAL_KEY = "return-val";

    private final static Object THROW_MARKER_EXCEPTION = 
        "throw-marker-exception";
    
    @Test
    private
    void
    testReflectUtilsInstantiatesInnerClasses()
        throws Exception
    {
        int i = 42;
        OuterClass oc = new OuterClass( i ); 

        OuterClass.InnerClass ic = 
            ReflectUtils.newInnerInstance( oc, OuterClass.InnerClass.class );
        state.equal( i, ic.getInt() );

        OuterClass.InnerStaticClass isc = 
            ReflectUtils.newInstance( OuterClass.InnerStaticClass.class );
        state.equal( OuterClass.STATIC_INT, isc.getInt() );
    }

    private
    final
    static
    class OuterClass
    {
        public final static int STATIC_INT = 7;

        private final int innerInt;

        OuterClass( int innerInt ) { this.innerInt = innerInt; }

        static
        class InnerStaticClass
        {
            public int getInt() { return STATIC_INT; }
        }

        final
        class InnerClass
        {
            public int getInt() { return innerInt; }
        }
    }

    @Test
    private
    void
    testGetResourceAsStream()
        throws Exception
    {
        InputStream is =
            ReflectUtils.getResourceAsStream( 
                LangTests.class, "test-resource" );

        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        byte[] buf = new byte[ 512 ];

        for ( int i = is.read( buf ); i >= 0; i = is.read( buf ) )
        {
            bos.write( buf, 0, i );
        }

        state.equalString(
            "Hello there, this was in a file\n",
            new String( bos.toByteArray(), "UTF-8" ) );
    }

    private
    Method
    findMethod( String nm )
    {
        Method res = null;
        
        for ( Method m : ReflectionTests.class.getDeclaredMethods() )
        {
            if ( m.getName().equals( nm ) )
            {
                if ( res == null ) res = m;
                else state.fail( "More than one method has name:", nm );
            }
        }

        state.isFalse( res == null, "No such method:", nm );
        res.setAccessible( true );

        return res;
    }

    private
    String
    instMethod0( Object markerVal,
                 String returnVal )
        throws Exception
    {
        return staticMethod0( markerVal, returnVal );
    }

    private
    static
    String
    staticMethod0( Object markerVal,
                   String returnVal )
        throws Exception
    {
        state.equal( MARKER_VAL1, markerVal );

        if ( returnVal.equals( THROW_MARKER_EXCEPTION ) )
        {
            throw new MarkerException();
        }
        else return returnVal;
    }

    static
    interface Foo
    {
        public
        void
        assertValue( String expct );
    }

    private
    final
    static
    class StaticFooBean
    implements Foo
    {
        private Object markerVal;
        private String returnVal;

        public
        void
        assertValue( String expct )
        {
            state.equalString( expct, returnVal );
        }
    }

    private
    final
    class InnerFooBean
    implements Foo
    {
        private Object markerVal;
        private String returnVal;

        public
        void
        assertValue( String expct )
        {
            state.equalString( expct, returnVal );
        }
    }

    // package level so it can be accessed in StandaloneFoo.java
    static
    abstract
    class AbstractFoo
    implements Foo
    {
        private final String returnVal;

        AbstractFoo( Object markerVal,
                     String returnVal ) 
            throws Exception
        { 
            staticMethod0( markerVal, returnVal );

            this.returnVal = returnVal;
        }

        public
        void
        assertValue( String expct )
        {
            state.equalString( expct, returnVal );
        }
    }

    private
    final
    static
    class StaticFoo
    extends AbstractFoo
    {
        private
        StaticFoo( Object markerVal,
                   String returnVal )
            throws Exception
        {
            super( markerVal, returnVal );
        }
    }

    private
    final
    class InnerFoo
    extends AbstractFoo
    {
        private
        InnerFoo( Object markerVal,
                  String returnVal )
            throws Exception
        {
            super( markerVal, returnVal );
        }
    }

    private
    Constructor< ? >
    firstConstructor( Class< ? > cls )
    {
        Constructor< ? > cons = cls.getDeclaredConstructors()[ 0 ];
        cons.setAccessible( true );

        return cons;
    }

    private
    < B extends AbstractMethodInvocation.Builder >
    B
    buildMethodInvocation0( B b,
                            Boolean ignoreUnmatchedKeys )
    {
        b.setKey( MARKER_VAL_KEY, 0 );
        b.setKey( RETURN_VAL_KEY, 1 );

        setIgnoreUnmatchedKeys( b, ignoreUnmatchedKeys );

        return b;
    }

    private
    void
    setInstance( AbstractInvocation.Builder< ?, ?, ? > b,
                 Class< ? > cls )
    {
        if ( cls.getEnclosingClass() != null &&
             ( ! ReflectUtils.isStatic( cls ) ) )
        {
            b.setInstance( this );
        }
    }

    private
    void
    setIgnoreUnmatchedKeys( AbstractInvocation.Builder< ?, ?, ? > b,
                            Boolean ignoreUnmatchedKeys )
    {
        if ( ignoreUnmatchedKeys != null )
        {
            b.setIgnoreUnmatchedKeys( ignoreUnmatchedKeys );
        }
    }

    private
    ReflectedInvocation
    getConstructor0Invocation( Class< ? > cls,
                               Boolean ignoreUnmatchedKeys )
    {
        Constructor< ? > cons = firstConstructor( cls );

        ConstructorInvocation.Builder b =
            new ConstructorInvocation.Builder().
                setTarget( cons );

        setInstance( b, cls );

        return buildMethodInvocation0( b, ignoreUnmatchedKeys ).build();
    }

    private
    static
    Field
    getField( Class< ? > cls,
              String nm )
        throws Exception
    {
        Field res = cls.getDeclaredField( nm );
        res.setAccessible( true );

        return res;
    }

    private
    final
    static
    class FooBeanPostProcessor
    implements BeanInvocation.PostProcessor
    {
        public
        Object
        process( Object val )
            throws Exception
        {
            Field f = val.getClass().getDeclaredField( "markerVal" );
            f.setAccessible( true );

            Object markerVal = 
                getField( val.getClass(), "markerVal" ).get( val );

            String returnVal = (String)
                getField( val.getClass(), "returnVal" ).get( val );
                
            staticMethod0( markerVal, returnVal );

            return val;
        }
    } 

    private
    BeanInvocation
    getBean0Invocation( Class< ? > cls,
                        Boolean ignoreUnmatchedKeys )
        throws Exception
    {
        BeanInvocation.Builder b =
            new BeanInvocation.Builder().
                setTargetAndConstructor( cls );

        setInstance( b, cls );
        setIgnoreUnmatchedKeys( b, ignoreUnmatchedKeys );

        b.setKey( MARKER_VAL_KEY, getField( cls, "markerVal" ) );
        b.setKey( RETURN_VAL_KEY, getField( cls, "returnVal" ) );

        b.setPostProcessor( new FooBeanPostProcessor() );

        return b.build();
    }

    private
    MethodInvocation
    getMethod0Invocation( String methName,
                          Object inst,
                          Boolean ignoreUnmatchedKeys )
    {
        MethodInvocation.Builder b = new MethodInvocation.Builder();
        b.setTarget( findMethod( methName ) );

        if ( inst != null ) b.setInstance( inst );

        buildMethodInvocation0( b, ignoreUnmatchedKeys );

        return b.build();
    }

    private
    Object
    call( ReflectedInvocation inv,
          Object... pairs )
        throws Exception
    {
        Map< Object, Object > params = 
            Lang.newMap( Object.class, Object.class, pairs );

        return inv.invoke( params );
    }

    @Test
    private
    void
    testInstMethod0SuccessStrictMatchingExplicit()
        throws Exception
    {
        Object res =
            call( getMethod0Invocation( "instMethod0", this, true ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello"
            );
 
        state.equalString( "hello", state.cast( String.class, res ) );
    }

    @Test
    private
    void
    testInstMethod0SuccessStrictMatchingImplicit()
        throws Exception
    {
        Object res =
            call( getMethod0Invocation( "instMethod0", this, null ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello"
            );
 
        state.equalString( "hello", state.cast( String.class, res ) );
    }

    @Test( expected = UnmatchedParameterKeyException.class,
           expectedPattern = "Unmatched parameter key: not-used" )
    private
    void
    testInstMethod0FailureOnStrictMatching()
        throws Exception
    {
        call( getMethod0Invocation( "instMethod0", this, null ),
              MARKER_VAL_KEY, MARKER_VAL1,
              RETURN_VAL_KEY, "hello",
              "not-used", new Object()
        );
    }

    @Test
    private
    void
    testInstMethod0SuccessIgnoreUnmatched()
        throws Exception
    {
        Object res = 
            call( getMethod0Invocation( "instMethod0", this, true ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello",
                  "not-used", new Object()
            );
 
        state.equalString( "hello", state.cast( String.class, res ) );
    }

    @Test( expected = MarkerException.class )
    private
    void
    testInstMethod0ExceptionType()
        throws Exception
    {
        Object res = 
            call( getMethod0Invocation( "instMethod0", this, null ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, THROW_MARKER_EXCEPTION
            );
    }

    // We only test a single static method for now rather than recover every
    // method already tested, since the main aim is to get basic coverage of the
    // null instance used for static methods. We can always expand this to cover
    // more if needed, but given the implementation a basic coverage of this
    // will suffice
    @Test
    private
    void
    testStaticMethod0SuccessStrictMatchingExplicit()
        throws Exception
    {
        Object res =
            call( getMethod0Invocation( "staticMethod0", null, true ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello"
            );
 
        state.equalString( "hello", state.cast( String.class, res ) );
    }

    private
    void
    assertFooSuccessStrictMatchingExplicit( Class< ? > fooCls,
                                            boolean isBean )
        throws Exception
    {
        ReflectedInvocation inv = isBean 
            ? getBean0Invocation( fooCls, true )
            : getConstructor0Invocation( fooCls, true );

        Foo res = (Foo)
            call( inv, MARKER_VAL_KEY, MARKER_VAL1, RETURN_VAL_KEY, "hello" );
 
        res.assertValue( "hello" );
    }

    private
    void
    testInnerFooSuccessStrictMatchingExplicit()
        throws Exception
    {
        assertFooSuccessStrictMatchingExplicit( InnerFoo.class, false );
    }
    
    @Test
    private
    void
    testInnerFooSuccessStrictMatchingImplicit()
        throws Exception
    {
        Foo res = (Foo)
            call( getConstructor0Invocation( InnerFoo.class, null ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello"
            );
 
        res.assertValue( "hello" );
    }

    @Test( expected = UnmatchedParameterKeyException.class,
           expectedPattern = "Unmatched parameter key: not-used" )
    private
    void
    testInnerFooFailureOnStrictMatching()
        throws Exception
    {
        Foo res = (Foo) 
            call( getConstructor0Invocation( InnerFoo.class, null ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello",
                  "not-used", new Object()
            );
 
        res.assertValue( "hello" );
    }

    @Test
    private
    void
    testInnerFooSuccessIgnoreUnmatched()
        throws Exception
    {
        Foo res = (Foo) 
            call( getConstructor0Invocation( InnerFoo.class, true ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello",
                  "not-used", new Object()
            );
 
        res.assertValue( "hello" );
    }

    @Test( expected = MarkerException.class )
    private
    void
    testInnerFooExceptionType()
        throws Exception
    {
        call( getConstructor0Invocation( InnerFoo.class, null ),
              MARKER_VAL_KEY, MARKER_VAL1,
              RETURN_VAL_KEY, THROW_MARKER_EXCEPTION
        );
    }

    // See note at testStaticMethod0SuccessStrictMatchingExplicit()
    @Test
    private
    void
    testStaticFooSuccessStrictMatchingExplicit()
        throws Exception
    {
        assertFooSuccessStrictMatchingExplicit( StaticFoo.class, false );
    }

    @Test
    private
    void
    testStandaloneFooStrictMatchingExplicit()
        throws Exception
    {
        assertFooSuccessStrictMatchingExplicit( StandaloneFoo.class, false );
    }
 
    @Test
    private
    void
    testInnerFooBeanSuccessStrictMatchingImplicit()
        throws Exception
    {
        Foo res = (Foo)
            call( getBean0Invocation( InnerFooBean.class, null ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello"
            );
 
        res.assertValue( "hello" );
    }

    @Test( expected = UnmatchedParameterKeyException.class,
           expectedPattern = "Unmatched parameter key: not-used" )
    private
    void
    testInnerFooBeanFailureOnStrictMatching()
        throws Exception
    {
        Foo res = (Foo) 
            call( getBean0Invocation( InnerFooBean.class, null ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello",
                  "not-used", new Object()
            );
 
        res.assertValue( "hello" );
    }

    @Test
    private
    void
    testInnerFooBeanSuccessIgnoreUnmatched()
        throws Exception
    {
        Foo res = (Foo) 
            call( getBean0Invocation( InnerFooBean.class, true ),
                  MARKER_VAL_KEY, MARKER_VAL1,
                  RETURN_VAL_KEY, "hello",
                  "not-used", new Object()
            );
 
        res.assertValue( "hello" );
    }

    @Test( expected = MarkerException.class )
    private
    void
    testInnerFooBeanExceptionType()
        throws Exception
    {
        call( getBean0Invocation( InnerFooBean.class, null ),
              MARKER_VAL_KEY, MARKER_VAL1,
              RETURN_VAL_KEY, THROW_MARKER_EXCEPTION
        );
    }

    // See note at testStaticMethod0SuccessStrictMatchingExplicit()
    @Test
    private
    void
    testStaticFooBeanSuccessStrictMatchingExplicit()
        throws Exception
    {
        assertFooSuccessStrictMatchingExplicit( StaticFooBean.class, true );
    }

    @Test
    private
    void
    testStandaloneFooBeanStrictMatchingExplicit()
        throws Exception
    {
        assertFooSuccessStrictMatchingExplicit( StandaloneFooBean.class, true );
    }

    private
    static
    void
    methodWithDefaults( byte aByte,
                        byte aDefaultByte, 
                        char aChar,
                        short aShort,
                        int anInt,
                        long aLong,
                        float aFloat,
                        double aDouble,
                        boolean aBoolean,
                        Object anObject,
                        Object aDefaultObject )
    {
        state.isTrue( (byte) 0 == aByte );
        state.isTrue( (byte) 3 == aDefaultByte );
        state.isTrue( (char) 0 == aChar );
        state.isTrue( (short) 0 == aShort );
        state.isTrue( 0 == anInt );
        state.isTrue( 0L == aLong );
        state.isTrue( 0.0f == aFloat );
        state.isTrue( 0.0d == aDouble );
        state.isFalse( aBoolean );
        state.isTrue( anObject == null );
        state.equal( MARKER_VAL1, aDefaultObject );

        code( "DONE" );
    }

    private
    void
    assertInvocationWithDefaults( 
        AbstractMethodInvocation.Builder< ?, ?, ? > b )
            throws Exception
    {
        ReflectedInvocation inv =
            b.setKey( "aByte", 0 ).
              setKey( "aDefaultByte", 1, Byte.valueOf( (byte) 3 ) ).
              setKey( "aChar", 2 ).
              setKey( "aShort", 3 ).
              setKey( "anInt", 4 ).
              setKey( "aLong", 5 ).
              setKey( "aFloat", 6 ).
              setKey( "aDouble", 7 ).
              setKey( "aBoolean", 8 ).
              setKey( "anObject", 9 ).
              setKey( "aDefaultObject", 10, MARKER_VAL1 ).
              build();
        
        inv.invoke( Lang.< Object, Object >emptyMap() );
    }

    @Test
    private
    void
    testMethodInvocationWithDefaults()
        throws Exception
    {
        assertInvocationWithDefaults(
            new MethodInvocation.Builder().
                setTarget( findMethod( "methodWithDefaults" ) ) );
    }

    private
    final
    static
    class DefaultConstructed
    {
        private
        DefaultConstructed( byte aByte,
                            byte aDefaultByte, 
                            char aChar,
                            short aShort,
                            int anInt,
                            long aLong,
                            float aFloat,
                            double aDouble,
                            boolean aBoolean,
                            Object anObject,
                            Object aDefaultObject )
        {
            methodWithDefaults( 
                aByte, aDefaultByte, aChar, aShort, anInt, aLong, aFloat,
                aDouble, aBoolean, anObject, aDefaultObject );
        }
    }

    @Test
    private
    void
    testConstructorInvocationWithDefaults()
        throws Exception
    {
        Constructor< ? > c = firstConstructor( DefaultConstructed.class );

        assertInvocationWithDefaults(
            new ConstructorInvocation.Builder().
                setTarget( firstConstructor( DefaultConstructed.class ) )
        );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = "No mapping for key: \\[NOT_EXPECTED\\]" )
    private
    void
    testAbstractMethodInvocationIndexOf()
        throws Exception
    {
        MethodInvocation mi = getMethod0Invocation( "instMethod0", this, true );

        state.equalInt( 0, mi.indexOf( MARKER_VAL_KEY ) );
        state.equalInt( 1, mi.indexOf( RETURN_VAL_KEY ) );

        mi.indexOf( "[NOT_EXPECTED]" );
    } 

    private final static class MarkerException extends Exception {}

    private
    static
    void
    method1()
    {}

    private
    static
    void
    method2( CharSequence p1,
             String p2 )
    {}

    private
    final
    static
    class Class1
    {
        private Class1() {}

        private
        Class1( CharSequence p1,
                String p2 )
        {}
    }

    private
    final
    static
    class CanCallContext
    {
        private Method m;
        private Constructor< ? > c;
    }

    private
    CanCallContext
    getCanCallContext( int indx )
        throws Exception
    {
        CanCallContext res = new CanCallContext();

        switch( indx )
        {
            case 1:
                res.m = getClass().getDeclaredMethod( "method1" );
                res.c = Class1.class.getDeclaredConstructor();
                break;
            
            case 2:
                Class< ? >[] typs = 
                    new Class< ? >[] { CharSequence.class, String.class };

                res.m = getClass().getDeclaredMethod( "method2", typs );
                res.c = Class1.class.getDeclaredConstructor( typs );
                break;
            
            default: state.fail();
        }

        return res;
    }

    private
    String
    completeErrExpct( int indx,
                      String errExpct,
                      boolean useMethod )
    {
        if ( errExpct == null ) return null;
        else
        {
            String pref = "Cannot call ";

            if ( useMethod ) 
            {
                return 
                    pref + "private static void com.bitgirder.lang.reflect." +
                    "ReflectionTests.method" + indx + errExpct;
            }
            else
            {
                return
                    pref + "private com.bitgirder.lang.reflect." +
                    "ReflectionTests$Class1" + errExpct;
            }
        }
    }

    private
    void
    assertCanCall( int indx,
                   CanCallContext ctx,
                   String errExpct,
                   Class< ? >[] typs,
                   boolean useMethod )
    {
        errExpct = completeErrExpct( indx, errExpct, useMethod );

        try
        {
            if ( useMethod ) ReflectUtils.assertCanCall( ctx.m, typs );
            else ReflectUtils.assertCanCall( ctx.c, typs );

            state.equal( null, errExpct );
        }
        catch ( IllegalStateException ise )
        {
            state.equalString( errExpct, ise.getMessage() );
        }
    }

    private
    void
    assertCanCall( int indx,
                   String errExpct,
                   Class< ? >... typs )
        throws Exception
    {
        CanCallContext ctx = getCanCallContext( indx );
        
        assertCanCall( indx, ctx, errExpct, typs, true );
        assertCanCall( indx, ctx, errExpct, typs, false );
    }

    @Test
    private
    void
    testCanCallMethod1Success()
        throws Exception
    {
        assertCanCall( 1, null );
    }

    @Test
    private
    void
    testCanCallMethod1Fails()
        throws Exception
    {
        assertCanCall( 
            1,
            "() with argument types: [class java.lang.Boolean]",
            Boolean.class 
        );
    }

    @Test
    private
    void
    testCanCallMethod2SuccessExact()
        throws Exception
    {
        assertCanCall( 
            2,
            null,
            CharSequence.class, String.class 
        );
    }

    @Test
    private
    void
    testCanCallMethod2SuccessAssignable()
        throws Exception
    {
        assertCanCall( 
            2,
            "(java.lang.CharSequence,java.lang.String) with argument types: [interface java.lang.CharSequence, class java.lang.StringBuilder]",
            CharSequence.class, StringBuilder.class 
        );
    }

    @Test
    private
    void
    testCanCallMethod2FailsTooManyArgs()
        throws Exception
    {
        assertCanCall( 
            2,
            "(java.lang.CharSequence,java.lang.String) with argument types: [interface java.lang.CharSequence, class java.lang.String, class java.lang.Integer]",
            CharSequence.class, String.class, Integer.class 
        );
    }

    @Test
    private
    void
    testCanCallMethod2FailsInvalidMatch()
        throws Exception
    {
        assertCanCall( 
            2,
            "(java.lang.CharSequence,java.lang.String) with argument types: [class java.lang.Integer, class java.lang.String]",
            Integer.class, String.class 
        );
    }

    @Test
    private
    void
    testCanCallMethod2SuperclassOfArgTypeFails()
        throws Exception
    {
        assertCanCall( 
            2,
            "(java.lang.CharSequence,java.lang.String) with argument types: [interface java.lang.CharSequence, interface java.lang.CharSequence]",
            CharSequence.class, CharSequence.class 
        );
    }

    final static class PkgFoo {}

    final void pkgFoo() {}

    @Test
    private
    void
    testHasPackageVisibility()
        throws Exception
    {
        state.isTrue( ReflectUtils.hasPackageVisibility( PkgFoo.class ) );
        state.isFalse( ReflectUtils.hasPackageVisibility( String.class ) );

        state.isTrue(
            ReflectUtils.hasPackageVisibility(
                ReflectionTests.class.getDeclaredMethod( "pkgFoo" ) ) );
        
        state.isFalse(
            ReflectUtils.hasPackageVisibility(
                ReflectionTests.class.
                    getDeclaredMethod( "testHasPackageVisibility" ) ) );
    }

    private
    static
    interface Function1
    {
        public
        int
        get()
            throws MarkerException;
    }

    private
    final
    static
    class Function1Impl
    implements Function1
    {
        private boolean called;

        private final int retVal;
        private final boolean doFail;

        private
        Function1Impl( int retVal,
                       boolean doFail )
        {
            this.retVal = retVal;
            this.doFail = doFail;
        }

        public
        int
        get()
            throws MarkerException
        {
            called = true;
            if ( doFail ) throw new MarkerException(); else return retVal;
        }
    }

    private
    Integer
    callProxy( boolean expctFail,
               Function1... funcs )
        throws MarkerException
    {
        Function1 f =
            ReflectUtils.
                createCallLoopProxy( Function1.class, Lang.asList( funcs ) );
        
        try 
        {
            int res = f.get();
            state.isFalse( expctFail, "get() returned:", res );
            return res;
        }
        catch ( MarkerException me ) 
        { 
            if ( expctFail ) return null; else throw me; 
        }
    }

    @Test
    private
    void
    testCallLoopProxyReturnVal()
        throws Exception
    {
        Function1Impl f1 = new Function1Impl( 1, false );
        Function1Impl f2 = new Function1Impl( 2, false );

        state.equalInt( 2, callProxy( false, f1, f2 ) );
        state.isTrue( f1.called );
        state.isTrue( f2.called );
    }

    @Test
    private
    void
    testCallLoopFirstElementProxyFailure()
        throws Exception
    {
        Function1Impl f1 = new Function1Impl( 1, true );
        Function1Impl f2 = new Function1Impl( 2, false );

        state.isTrue( callProxy( true, f1, f2 ) == null );
        state.isTrue( f1.called );
        state.isFalse( f2.called );

        f1 = new Function1Impl( 1, false );
        f2 = new Function1Impl( 2, true );
        state.isTrue( callProxy( true, f1, f2 ) == null );
        state.isTrue( f1.called );
        state.isTrue( f2.called );
    }
}
