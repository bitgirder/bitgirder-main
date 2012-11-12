package com.bitgirder.mingle.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.mingle.codec.MingleDecoder;

import com.bitgirder.io.AbstractObjectLineProcessor;

import java.util.List;

import java.nio.ByteBuffer;

public
final
class JsonMingleStructLineProcessor
extends AbstractObjectLineProcessor< 
            JsonMingleStructLineProcessor.MingleStructParser, 
            MingleStruct >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final JsonMingleCodec codec = JsonMingleCodecs.getJsonCodec();

    private JsonMingleStructLineProcessor( Builder b ) { super( b ); }

    protected
    MingleStructParser
    createParser( String fileName )
    {
        return 
            new MingleStructParser( codec.createDecoder( MingleStruct.class ) );
    }

    protected
    void
    update( MingleStructParser p,
            ByteBuffer bb,
            boolean isEnd )
        throws Exception
    {
        p.update( bb, isEnd );
    }

    protected
    MingleStruct
    buildParseResult( MingleStructParser p )
        throws Exception
    {
        return p.dec.getResult();
    }

    // making it package-accessible so it can be referenced as type param in
    // this class's definition
    final
    static
    class MingleStructParser
    {
        private final MingleDecoder< MingleStruct > dec;

        private boolean done;

        private
        MingleStructParser( MingleDecoder< MingleStruct > dec )
        {
            this.dec = dec;
        }

        private void noOpConsume( ByteBuffer bb ) { bb.position( bb.limit() ); }

        private
        void
        update( ByteBuffer bb,
                boolean isEnd )
            throws Exception
        {
            if ( done ) noOpConsume( bb );
            else
            {
                done = dec.readFrom( bb, isEnd );
                if ( done ) noOpConsume( bb );
            }
        }
    }

    public
    final
    static
    class Builder
    extends AbstractObjectLineProcessor.Builder< MingleStructParser,
                                                 MingleStruct,
                                                 Builder >
    {
        public
        JsonMingleStructLineProcessor
        build()
        {
            return new JsonMingleStructLineProcessor( this );
        }
    }
}
