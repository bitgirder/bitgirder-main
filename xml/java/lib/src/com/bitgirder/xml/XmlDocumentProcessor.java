package com.bitgirder.xml;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.AbstractProtocolProcessor;

import java.io.ByteArrayOutputStream;

import java.nio.ByteBuffer;

import org.w3c.dom.Document;

public
final
class XmlDocumentProcessor
extends AbstractProtocolProcessor< ByteBuffer >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final byte[] xferBuf = new byte[ 512 ];
    private final ByteArrayOutputStream bos = new ByteArrayOutputStream();

    private Document doc;

    public
    Document
    getDocument()
    {
        state.isFalse( 
            doc == null, "Attempt to access document before completion" );
 
        return doc;
    }

    private
    void
    parseDoc()
        throws Exception
    {
        bos.close();
        doc = XmlIo.parseDocument( bos.toByteArray() );
    }

    protected
    void
    processImpl( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        while ( ctx.object().hasRemaining() )
        {
            int len = Math.min( xferBuf.length, ctx.object().remaining() );
            ctx.object().get( xferBuf, 0, len );

            bos.write( xferBuf, 0, len );
        }

        if ( ctx.isFinal() ) parseDoc();

        doneOrData( ctx );
    }

    public
    static
    XmlDocumentProcessor
    create()
    {
        return new XmlDocumentProcessor(); 
    }
}
