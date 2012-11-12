package com.bitgirder.aws;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.crypto.CryptoUtils;

import com.bitgirder.io.FileWrapper;

import java.io.FileInputStream;
import java.io.IOException;

import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.UnrecoverableEntryException;

import java.security.cert.CertificateException;

import javax.crypto.SecretKey;

public
final
class AwsKeyUtils
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private AwsKeyUtils() {}

    private
    static
    KeyStore
    loadKeyStore( KeyStoreLookup lookup )
        throws IOException,
               NoSuchAlgorithmException,
               KeyStoreException,
               CertificateException
    {
        KeyStore res = KeyStore.getInstance( lookup.keyStoreType );

        char[] pass = 
            lookup.keyStorePass == null
                ? CryptoUtils.readPassword( "KeyStore password: " )
                : lookup.keyStorePass;

        FileInputStream is = 
            new FileInputStream( lookup.keyStoreFile.getFile() );

        try { res.load( is, pass ); } finally { is.close(); }
 
        return res;
    }

    public
    static
    SecretKey
    getAwsSecretKey( KeyStoreLookup lookup )
        throws IOException,
               NoSuchAlgorithmException,
               KeyStoreException,
               CertificateException,
               UnrecoverableEntryException
    {
        inputs.notNull( lookup, "lookup" );

        KeyStore ks = loadKeyStore( lookup );
 
        char[] pass = 
            lookup.keyPass == null
                ? CryptoUtils.readPassword( "AWS secret key password: " )
                : lookup.keyPass;

        KeyStore.ProtectionParameter pp =
            new KeyStore.PasswordProtection( pass );

        KeyStore.SecretKeyEntry keyEntry =
            (KeyStore.SecretKeyEntry) ks.getEntry( lookup.keyAlias, pp );
 
        return keyEntry.getSecretKey();
    }
 
    public
    final
    static
    class KeyStoreLookup
    {
        private final FileWrapper keyStoreFile; // not null
        private final char[] keyStorePass; // could be null

        // can make configurable later
        private final String keyStoreType = "jceks"; 
    
        private final String keyAlias; // not null
        private final char[] keyPass; // could be null

        private
        KeyStoreLookup( Builder b )
        {
            this.keyStoreFile = 
                inputs.notNull( b.keyStoreFile, "keyStoreFile" );

            this.keyStorePass = b.keyStorePass;
    
            this.keyAlias = inputs.notNull( b.keyAlias, "keyAlias" );
            this.keyPass = b.keyPass;
        }
 
        // setters which take passwords as char[] do not copy, so callers can
        // zero passwords out at any time to invalidate any lookups built from
        // this builder. Note that doing so will not lead to useful
        // error messages if a program does attempt to use such a Lookup after
        // a password is zeroed out.
        //
        // Passwords do not need to be set at all; methods in AwsKeyUtils which
        // need passwords but do not find them will use the system console to
        // prompt for them
        public
        final
        static
        class Builder
        {
            private FileWrapper keyStoreFile;
            private char[] keyStorePass;
    
            private String keyAlias;
            private char[] keyPass;
    
            public
            Builder
            setKeyStoreFile( FileWrapper keyStoreFile )
            {
                this.keyStoreFile = 
                    inputs.notNull( keyStoreFile, "keyStoreFile" );

                return this;
            }

            public
            Builder
            setKeyStoreFile( CharSequence keyStoreFile )
            {
                return setKeyStoreFile(
                    new FileWrapper( 
                        inputs.notNull( keyStoreFile, "keyStoreFile" ) ) );
            }

            public
            Builder
            setKeyStorePass( char[] keyStorePass )
            {
                this.keyStorePass = 
                    inputs.notNull( keyStorePass, "keyStorePass" );
                
                return this;
            }
    
            public
            Builder
            setKeyStorePass( String keyStorePass )
            {
                inputs.notNull( keyStorePass, "keyStorePass" );
                return setKeyStorePass( keyStorePass.toCharArray() );
            }
    
            public
            Builder
            setKeyAlias( String keyAlias )
            {
                this.keyAlias = inputs.notNull( keyAlias, "KeyAlias" );
                return this;
            }

            public
            Builder
            setKeyPass( char[] keyPass )
            {
                this.keyPass = inputs.notNull( keyPass, "KeyPass" );
                return this;
            }
    
            public
            Builder
            setKeyPass( String keyPass )
            {
                inputs.notNull( keyPass, "KeyPass" );
                return setKeyPass( keyPass.toCharArray() );
            }

            public KeyStoreLookup build() { return new KeyStoreLookup( this ); }
        }
    }
}
