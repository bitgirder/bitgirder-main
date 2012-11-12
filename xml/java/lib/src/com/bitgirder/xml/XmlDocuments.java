package com.bitgirder.xml;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;

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

    static
    {
        dbf = DocumentBuilderFactory.newInstance();
        dbf.setNamespaceAware( true );
    }
}
