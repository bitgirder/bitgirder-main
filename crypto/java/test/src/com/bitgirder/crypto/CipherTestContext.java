package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;

public
final
class CipherTestContext
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String trans;
    private final int keyLen;

    private
    CipherTestContext( String trans,
                       int keyLen )
    {
        this.trans = trans;
        this.keyLen = keyLen;
    }

    final String transformation() { return trans; }
    final int keyLen() { return keyLen; }

    public
    CharSequence
    getLabel()
    {
        return
            Strings.crossJoin( "=", ",",
                "trans", trans,
                "keyLen", keyLen
            );
    }

    public
    static
    CipherTestContext
    create( String trans,
            int keyLen )
    {
        inputs.notNull( trans, "trans" );
        inputs.positiveI( keyLen, "keyLen" );

        return new CipherTestContext( trans, keyLen );
    }
}
