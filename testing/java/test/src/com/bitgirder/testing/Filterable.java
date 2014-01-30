package com.bitgirder.testing;

import com.bitgirder.lang.Lang;

import com.bitgirder.test.Test;
import com.bitgirder.test.Tests;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import java.util.List;

@Test
final
class Filterable
{
    @Test private void test1() {}
    @Test private void test2() {}

    private
    final
    class CallImpl
    extends LabeledTestCall
    {
        private CallImpl( String nm ) { super( nm ); }

        public void call() {}
    }

    @InvocationFactory
    private
    List< CallImpl >
    testCall()
    {
        List< CallImpl > res = Lang.newList();

        res.add( new CallImpl( "call1" ) );
        res.add( new CallImpl( "call2" ) );

        return res;
    }

    public
    final
    static
    class InstanceTest
    implements LabeledTestObject
    {
        private final int i;

        private InstanceTest( int i ) { this.i = i; }

        @Test public void test1() {}
        @Test public void test2() {}

        public String getLabel() { return "inst" + i; }

        public Object getInvocationTarget() { return this; }
    }

    @TestFactory
    private
    List< InstanceTest >
    instFactory()
    {
        List< InstanceTest > res = Lang.newList();

        res.add( new InstanceTest( 1 ) );
        res.add( new InstanceTest( 2 ) );

        return res;
    }

    @TestFactory
    private
    static
    List< LabeledTestObject >
    factory()
    {
        List< LabeledTestObject > res = Lang.newList();

        res.add( Tests.createLabeledTestObject( new Filterable(), "static1" ) );
        res.add( Tests.createLabeledTestObject( new Filterable(), "static2" ) );

        return res;
    }
}
