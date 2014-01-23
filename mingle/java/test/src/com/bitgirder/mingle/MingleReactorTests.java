package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.pipeline.Pipelines;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import java.util.List;

@Test
final
class MingleReactorTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static DeclaredTypeName TYP_VALUE_BUILD_TEST =
        DeclaredTypeName.create( "ValueBuildTest" );

    private static void code( Object... args ) { CodeLoggers.code( args ); }

    private 
    static 
    void 
    codef( String tmpl, 
           Object... args ) 
    { 
        CodeLoggers.codef( tmpl, args ); 
    }

    private
    static
    abstract
    class TestImpl
    extends LabeledTestCall
    {
        private TestImpl( CharSequence nm ) { super( nm ); }
    }

    private
    static
    final
    class ValueBuildTest
    extends TestImpl
    {
        private MingleValue val;

        private ValueBuildTest( CharSequence nm ) { super( nm ); }

        public
        void
        call()
            throws Exception
        {
            MingleValueReactorPipeline pip = 
                MingleValueReactors.createValueBuilderPipeline();

            MingleValueReactors.visitValue( val, pip );

            MingleValueBuilder bld = 
                Pipelines.lastElementOfType( 
                    pip.pipeline(), MingleValueBuilder.class );
            
            MingleTests.assertEqual( val, bld.value() );
        }
    }

    private
    static
    class TestImplReader
    extends MingleTestGen.StructFileReader< TestImpl >
    {
        private
        TestImplReader()
        {
            super( "reactor-tests.bin" );
        }

        private
        CharSequence
        makeName( MingleStructAccessor testObj,
                  CharSequence name )
        {
            return testObj.getType().getName() + "/" + name;
        }

        private
        ValueBuildTest
        convertValueBuildTest( MingleStructAccessor acc )
        {
            MingleValue val = acc.expectMingleValue( "val" );

            String nm = String.format( "%s (%s)", 
                Mingle.inspect( val ), val.getClass().getName() );

            ValueBuildTest res = new ValueBuildTest( makeName( acc, nm ) );
            res.val = val;

            return res;
        }

        protected
        TestImpl
        convertStruct( MingleStruct ms )
        {
            DeclaredTypeName nm = ms.getType().getName();

            MingleStructAccessor acc = MingleStructAccessor.forStruct( ms );

            if ( nm.equals( TYP_VALUE_BUILD_TEST ) ) {
                return convertValueBuildTest( acc );
            }
            
            codef( "skipping test: %s", nm );
            return null;
        }
    }

    @InvocationFactory
    private
    List< TestImpl >
    testReactor()
        throws Exception
    {
        return new TestImplReader().read();
    }
}
