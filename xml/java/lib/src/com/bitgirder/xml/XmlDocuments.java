package com.bitgirder.xml;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.InputStream;

import javax.xml.validation.Schema;
import javax.xml.validation.SchemaFactory;

import javax.xml.XMLConstants;

import javax.xml.transform.stream.StreamSource;

import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;

import org.xml.sax.SAXException;

import org.w3c.dom.Document;

public
final
class XmlDocuments
{
    private final static DocumentBuilderFactory dbf;

    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private XmlDocuments() {}

    public
    static
    DocumentBuilder
    newDocumentBuilder()
    {
        try { return dbf.newDocumentBuilder(); }
        catch ( Throwable th )
        {
            throw 
                new RuntimeException( "Couldn't create document builder", th );
        }
    }

    public
    static
    Document
    newDocument()
    {
        return newDocumentBuilder().newDocument();
    }

    public
    static
    Schema
    loadSchema( InputStream is )
        throws SAXException
    {
        inputs.notNull( is, "is" );

        SchemaFactory fact = 
            SchemaFactory.newInstance( XMLConstants.W3C_XML_SCHEMA_NS_URI );

        StreamSource ss = new StreamSource( is );
        return fact.newSchema( ss );
    }

    static
    {
        dbf = DocumentBuilderFactory.newInstance();
        dbf.setNamespaceAware( true );
    }
}
