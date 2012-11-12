package com.bitgirder.xml;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import javax.xml.xpath.XPathFactory;
import javax.xml.xpath.XPath;
import javax.xml.xpath.XPathExpressionException;

import org.w3c.dom.Node;

public
final
class Xpaths
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static XPathFactory xpf = XPathFactory.newInstance();

    private final static ThreadLocal< XPath > localXPath = 
        new ThreadLocal< XPath >() {
            @Override protected XPath initialValue() { return xpf.newXPath(); }
        };

    private Xpaths() {}

    private static XPath getXPath() { return localXPath.get(); }

    public
    static
    String
    evaluate( String expr,
              Node n )
        throws XPathExpressionException
    {
        inputs.notNull( expr, "expr" );
        inputs.notNull( n, "n" );

        return getXPath().evaluate( expr, n );
    }
}
