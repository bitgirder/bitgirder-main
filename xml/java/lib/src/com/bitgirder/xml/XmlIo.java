package com.bitgirder.xml;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolProcessors;

import java.io.ByteArrayOutputStream;
import java.io.ByteArrayInputStream;

import java.util.Iterator;

import java.nio.ByteBuffer;

import javax.xml.parsers.DocumentBuilder;

import javax.xml.transform.Transformer;
import javax.xml.transform.TransformerFactory;

import javax.xml.transform.dom.DOMSource;

import javax.xml.transform.stream.StreamResult;

import org.w3c.dom.Document;

public
final
class XmlIo
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static TransformerFactory tfact =
        TransformerFactory.newInstance();

    private XmlIo() {}

    public
    static
    Document
    parseDocument( byte[] arr,
                   int off,
                   int len )
        throws Exception
    {
        inputs.notNull( arr, "arr" );

        DocumentBuilder db = XmlDocuments.newDocumentBuilder();

        ByteArrayInputStream bis = new ByteArrayInputStream( arr, off, len );

        try { return db.parse( bis ); } finally { bis.close(); }
    }

    public
    static
    Document
    parseDocument( byte[] arr )
        throws Exception
    {
        inputs.notNull( arr, "arr" );
        return parseDocument( arr, 0, arr.length );
    }

    public
    static
    Document
    parseDocument( Iterable< ? extends ByteBuffer > bufs )
        throws Exception
    {
        inputs.notNull( bufs, "bufs" );

        XmlDocumentProcessor proc = XmlDocumentProcessor.create();

        int indx = 0;

        for ( Iterator< ? extends ByteBuffer > it = bufs.iterator();
                it.hasNext(); )
        {
            ByteBuffer bb = it.next();

            inputs.isFalse( 
                bb == null, "Buffer", ++indx, "in iterator is null" );
            
            ProtocolProcessors.processImmediate( proc, bb, ! it.hasNext() );
        }

        return proc.getDocument();
    }

    public
    static
    byte[]
    toByteArray( Document doc )
        throws Exception
    {
        inputs.notNull( doc, "doc" );

        ByteArrayOutputStream bos = new ByteArrayOutputStream();

        tfact.newTransformer().
              transform( new DOMSource( doc ), new StreamResult( bos ) );
        
        bos.close();
        return bos.toByteArray();
    }

    public
    static
    ByteBuffer
    toByteBuffer( Document doc )
        throws Exception
    {
        return ByteBuffer.wrap( toByteArray( doc ) );
    }

    public
    static
    CharSequence
    toString( Document doc )
        throws Exception
    {
        inputs.notNull( doc, "doc" );

        String enc = doc.getXmlEncoding();
        inputs.isFalse( enc == null, "Document has no XML Encoding specified" );

        return new String( toByteArray( doc ), enc );
    }
}
