package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleBinaryConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.BinWriter;

import java.util.List;

import java.io.OutputStream;
import java.io.IOException;

public
final
class MingleBinWriter
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final BinWriter w;

    private MingleBinWriter( BinWriter w ) { this.w = w; }

    private
    void
    writeTypeCode( byte tc )
        throws IOException
    {
        w.writeByte( tc );
    }

    private
    void
    writeUint8( int i )
        throws IOException
    {
        w.writeByte( Lang.fromOctet( i ) );
    }

    public
    void
    writeIdentifier( MingleIdentifier id )
        throws IOException
    {
        inputs.notNull( id, "id" );

        writeTypeCode( TC_ID );

        String[] parts = id.getPartsArray();

        writeUint8( parts.length );
        for ( String part : parts ) w.writeUtf8( part );
    }

    private
    void
    writeIdentifiers( List< MingleIdentifier > ids )
        throws IOException
    {
        writeUint8( ids.size() );
        for ( MingleIdentifier id : ids ) writeIdentifier( id );
    }

    public
    void
    writeDeclaredTypeName( DeclaredTypeName nm )
        throws IOException
    {
        inputs.notNull( nm, "nm" );

        writeTypeCode( TC_DECL_NM );
        w.writeUtf8( nm.getExternalForm() );
    }

    public
    void
    writeNamespace( MingleNamespace ns )
        throws IOException
    {
        writeTypeCode( TC_NS );
        
        writeIdentifiers( ns.getParts() );
        writeIdentifier( ns.getVersion() );
    }

    public
    void
    writeQualifiedTypeName( QualifiedTypeName qn )
        throws IOException
    {
        inputs.notNull( qn, "qn" );

        writeTypeCode( TC_QN );
        writeNamespace( qn.getNamespace() );
        writeDeclaredTypeName( qn.getName() );
    }

    public
    static
    MingleBinWriter
    create( OutputStream os )
    {
        inputs.notNull( os, "os" );
        return new MingleBinWriter( BinWriter.asWriterLe( os ) );
    }
}
