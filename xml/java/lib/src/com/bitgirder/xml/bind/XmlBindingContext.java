package com.bitgirder.xml.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.xml.XmlDocuments;

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
    byte[]
    toByteArray( Object obj )
    {
        inputs.notNull( obj, "obj" );

        try
        {
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            getMarshaller().marshal( obj, bos );

            return bos.toByteArray();
        }
        catch ( Throwable th )
        { 
            throw createRethrow( "Couldn't marshal object", th );
        }
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
    fromByteArray( byte[] arr,
                   Class< V > cls )
    {
        inputs.notNull( arr, "arr" );
        inputs.notNull( cls, "cls" );

        try
        {
            ByteArrayInputStream bis = new ByteArrayInputStream( arr );
            StreamSource src = new StreamSource( bis );

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
