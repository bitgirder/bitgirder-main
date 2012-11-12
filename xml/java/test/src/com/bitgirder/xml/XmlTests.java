package com.bitgirder.xml;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.Charsets;

import com.bitgirder.test.Test;

import java.util.List;

import java.nio.ByteBuffer;

import org.w3c.dom.Document;

@Test
final
class XmlTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static String STR1 = "String1";

    private final static String DOC1_STR =
        "<?xml version=\"1.0\" encoding=\"utf-8\"?>" +
        "<doc1><child>" + STR1 + "</child></doc1>";

    private
    byte[]
    getDoc1Bytes()
        throws Exception
    {
        return DOC1_STR.getBytes( Charsets.UTF_8.charset() );
    }

    private
    void
    assertDoc1( Document doc )
        throws Exception
    {
        state.equalString( STR1, Xpaths.evaluate( "/doc1/child/text()", doc ) );
    }

    private
    Document
    parseDoc1()
        throws Exception
    {
        return XmlIo.parseDocument( getDoc1Bytes() );
    }

    @Test
    private
    void
    testXmlIoParseDocFromByteArray()
        throws Exception
    {
        assertDoc1( XmlIo.parseDocument( getDoc1Bytes() ) );
    }

    @Test
    private
    void
    testXmlIoParseDocFromBufferList()
        throws Exception
    {
        ByteBuffer bb = ByteBuffer.wrap( getDoc1Bytes() );
        List< ByteBuffer > l = Lang.newList();
        
        for ( int i = 0, e = bb.limit(); i < e; i += 3 )
        {
            ByteBuffer bb2 = bb.duplicate();
            bb2.position( i );
            bb2.limit( Math.min( i + 3, bb.limit() ) );

            l.add( bb2 );
        }

        state.isTrue( l.size() > 1 ); // sanity check that test is interesting
        assertDoc1( XmlIo.parseDocument( l ) );
    }

    @Test
    private
    void
    testXmlDocumentProcessorBasic()
        throws Exception
    {
        XmlDocumentProcessor proc = XmlDocumentProcessor.create();

        ProtocolProcessors.
            processImmediate( 
                proc, 
                ByteBuffer.wrap( getDoc1Bytes() ), 
                true 
            );

        assertDoc1( proc.getDocument() );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = "Attempt to access document before completion" )
    private
    void
    testPrematureDocumentAccessFails()
    {
        XmlDocumentProcessor.create().getDocument();
    }

    @Test
    private
    void
    testDocumentToString()
        throws Exception
    {
        Document doc = parseDoc1();
        doc.setXmlStandalone( true );

        CharSequence doc1Str = XmlIo.toString( doc );
        state.equalString( DOC1_STR, doc1Str );
    }
}
