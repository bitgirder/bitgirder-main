package com.bitgirder.xml.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.xml.XmlDocuments;

import java.io.OutputStream;
import java.io.InputStream;
import java.io.ByteArrayOutputStream;
import java.io.ByteArrayInputStream;

import javax.xml.transform.stream.StreamSource;

import javax.xml.bind.JAXBContext;
import javax.xml.bind.JAXBException;
import javax.xml.bind.JAXBElement;
import javax.xml.bind.Marshaller;
import javax.xml.bind.Unmarshaller;

import org.w3c.dom.Document;

public
final
class XmlBindingContext
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final JAXBContext jbCtx;

    private XmlBindingContext( JAXBContext jbCtx ) { this.jbCtx = jbCtx; }

    private
    static
    RuntimeException
    createRethrow( String msg,
                   Throwable th )
    {
        return new RuntimeException( msg, th );
    }

    public
    Marshaller
    getMarshaller()
    {
        try { return jbCtx.createMarshaller(); }
        catch ( Throwable th )
        {
            throw createRethrow( "Couldn't create marshaller", th );
        }
    }

    public
    void
    writeObject( Object obj,
                 OutputStream os )
    {
        inputs.notNull( obj, "obj" );
        inputs.notNull( os, "os" );
        
        try { getMarshaller().marshal( obj, os ); }
        catch ( Throwable th )
        { 
            throw createRethrow( "Couldn't marshal object", th );
        }
    }

    public
    byte[]
    toByteArray( Object obj )
    {
        inputs.notNull( obj, "obj" );

        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        writeObject( obj, bos );

        return bos.toByteArray();
    }

    public
    Document
    toDocument( Object obj )
    {
        inputs.notNull( obj, "obj" );

        try
        {
            Document doc = XmlDocuments.newDocument();
            getMarshaller().marshal( obj, doc );

            return doc;
        }
        catch ( Throwable th )
        {
            throw createRethrow( "Couldn't marshal object", th );
        }
    }

    public
    Unmarshaller
    getUnmarshaller()
    {
        try { return jbCtx.createUnmarshaller(); }
        catch ( Throwable th )
        {
            throw createRethrow( "Couldn't create unmarshaller", th );
        }
    }

    public
    < V >
    V
    readObject( InputStream is,
                Class< V > cls )
    {
        inputs.notNull( is, "is" );
        inputs.notNull( cls, "cls" );

        try
        {
            StreamSource src = new StreamSource( is );
            return getUnmarshaller().unmarshal( src, cls ).getValue();
        }
        catch ( Throwable th ) 
        {
            throw createRethrow( "Couldn't unmarshal input", th );
        }
    }

    public
    < V >
    V
    fromByteArray( byte[] arr,
                   Class< V > cls )
    {
        inputs.notNull( arr, "arr" );
        inputs.notNull( cls, "cls" );

        ByteArrayInputStream bis = new ByteArrayInputStream( arr );
        return readObject( bis, cls );
    }

    public
    < V >
    V
    fromDocument( Document doc,
                  Class< V > cls )
    {
        inputs.notNull( doc, "doc" );
        inputs.notNull( cls, "cls" );

        try { return getUnmarshaller().unmarshal( doc, cls ).getValue(); }
        catch ( Throwable th )
        {
            throw createRethrow( "Couldn't unmarshal doc", th );
        }
    }

    public
    static
    XmlBindingContext
    create( JAXBContext jbCtx )
    {
        return new XmlBindingContext( inputs.notNull( jbCtx, "jbCtx" ) );
    }

    public
    static
    XmlBindingContext
    create( String ctxPath )
    {
        inputs.notNull( ctxPath, "ctxPath" );
        try { return create( JAXBContext.newInstance( ctxPath ) ); }
        catch ( Throwable th )
        {
            throw 
                createRethrow( 
                    "Couldn't create context for path: " + ctxPath, th );
        }
    }
}
