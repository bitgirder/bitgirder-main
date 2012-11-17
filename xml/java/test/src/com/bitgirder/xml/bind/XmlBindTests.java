package com.bitgirder.xml.bind;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.xml.XmlIo;

import com.bitgirder.test.Test;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;

import org.w3c.dom.Document;

@Test
final
class XmlBindTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private
    XmlBindingContext
    context()
    {
        return XmlBindingContext.create( "com.bitgirder.xml.bind" );
    }

    private ObjectFactory objectFactory() { return new ObjectFactory(); }

    private
    Struct1
    createStruct1Inst1()
    {
        Struct1 res = objectFactory().createStruct1();
        res.setAString( "hello" );

        return res;
    }

    private
    void
    assertEqual( Struct1 s1,
                 Struct1 s2 )
    {
        if ( state.sameNullity( s1, s2 ) )
        {
            state.equal( s1.getAString(), s2.getAString() );
        }
    }

    @Test
    private
    void
    testBinaryRoundtrip()
        throws Exception
    {
        Struct1 s1 = createStruct1Inst1();

        byte[] docArr = context().toByteArray( s1 );
        code( "doc:", new String( docArr, "UTF-8" ) );

        Struct1 s2 = context().fromByteArray( docArr, Struct1.class );
        assertEqual( s1, s2 );
    }

    @Test
    private
    void
    testStreamRoundtrip()
        throws Exception
    {
        Struct1 s1 = createStruct1Inst1();

        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        context().writeObject( s1, bos );

        ByteArrayInputStream bis = 
            new ByteArrayInputStream( bos.toByteArray() );
        
        Struct1 s2 = context().readObject( bis, Struct1.class );
        assertEqual( s1, s2 );
    }

    @Test
    private
    void
    testDocumentRoundtrip()
        throws Exception
    {
        Struct1 s1 = createStruct1Inst1();

        Document doc = context().toDocument( s1 );

        Struct1 s2 = context().fromDocument( doc, Struct1.class );
        assertEqual( s1, s2 );
    }
}
