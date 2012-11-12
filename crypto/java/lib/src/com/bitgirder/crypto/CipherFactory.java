package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.security.Key;
import java.security.GeneralSecurityException;

import java.security.spec.AlgorithmParameterSpec;

import javax.crypto.Cipher;

import javax.crypto.spec.IvParameterSpec;

public
final
class CipherFactory
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String trans;
    private final Key key;
    private final AlgorithmParameterSpec aps;

    private
    CipherFactory( Builder b )
    {
        this.trans = inputs.notNull( b.trans, "trans" );
        this.key = inputs.notNull( b.key, "key" );
        this.aps = b.aps;
    }

    public String getTransformation() { return trans; }
    public Key getKey() { return key; }
    public AlgorithmParameterSpec getAlgorithmParameterSpec() { return aps; }

    public
    Cipher
    initCipher( int opmode )
        throws GeneralSecurityException
    {
        Cipher res = CryptoUtils.createCipher( trans );

        if ( aps == null ) CryptoUtils.initCipher( opmode, res, key );
        else CryptoUtils.initCipher( opmode, res, key, aps );

        return res;
    }

    public 
    Cipher 
    initEncrypt() 
        throws GeneralSecurityException
    { 
        return initCipher( Cipher.ENCRYPT_MODE ); 
    }

    public 
    Cipher 
    initDecrypt() 
        throws GeneralSecurityException
    { 
        return initCipher( Cipher.DECRYPT_MODE ); 
    }

    public
    final
    static
    class Builder
    {
        private String trans;
        private Key key;
        private AlgorithmParameterSpec aps;

        public
        Builder
        setTransformation( String trans )
        {
            this.trans = inputs.notNull( trans, "trans" );
            return this;
        }

        public
        Builder
        setKey( Key key )
        {
            this.key = inputs.notNull( key, "key" );
            return this;
        }

        public
        Builder
        setAlgorithmParameterSpec( AlgorithmParameterSpec aps )
        {
            this.aps = inputs.notNull( aps, "aps" );
            return this;
        }

        public
        Builder
        setIv( IvParameterSpec iv )
        {
            return setAlgorithmParameterSpec( inputs.notNull( iv, "iv" ) );
        }

        public CipherFactory build() { return new CipherFactory( this ); }
    }
}
