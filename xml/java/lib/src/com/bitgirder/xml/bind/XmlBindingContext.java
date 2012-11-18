package com.bitgirder.xml.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.xml.XmlDocuments;

import java.io.OutputStream;
import java.io.InputStream;
import java.io.ByteArrayOutputStream;
import java.io.ByteArrayInputStream;

import javax.xml.transform.stream.StreamSource;

import javax.xml.validation.Schema;

import javax.xml.bind.JAXBContext;
import javax.xml.bind.JAXBException;
import javax.xml.bind.JAXBElement;
import javax.xml.bind.Marshaller;
import javax.xml.bind.Unmarshaller;

import org.w3c.dom.Document;

// Callers which mutate an instance, via methods such as setSchema, must manage
// their own expectations on concurrency and ensure, for instance, that calls to
// setSchema() have a happens-before relationship to any other calls which might
// use that schema.
public
final
class XmlBindingContext
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final JAXBContext jbCtx;

    private Schema schema;

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
        throws JAXBException
    {
        Unmarshaller res = jbCtx.createUnmarshaller();

        if ( schema != null ) res.setSchema( schema );

        return res;
    }

    public
    < V >
    V
    readObject( InputStream is,
                Class< V > cls )
        throws JAXBException
    {
        inputs.notNull( is, "is" );
        inputs.notNull( cls, "cls" );

        StreamSource src = new StreamSource( is );
        return getUnmarshaller().unmarshal( src, cls ).getValue();
    }

    public
    < V >
    V
    fromByteArray( byte[] arr,
                   Class< V > cls )
        throws JAXBException
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
        throws JAXBException
    {
        inputs.notNull( doc, "doc" );
        inputs.notNull( cls, "cls" );

        return getUnmarshaller().unmarshal( doc, cls ).getValue();
    }

    public
    void
    setSchema( Schema s )
    {
        inputs.notNull( s, "s" );
        this.schema = s;
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
