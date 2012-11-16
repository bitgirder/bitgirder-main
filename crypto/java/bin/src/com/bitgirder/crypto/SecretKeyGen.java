package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.FileWrapper;

import com.bitgirder.application.ApplicationProcess;

import java.io.FileOutputStream;
import java.io.FileInputStream;
import java.io.Console;

import java.util.Enumeration;

import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.SecureRandom;

import javax.crypto.KeyGenerator;
import javax.crypto.SecretKey;

// TODO: add a backup mechanism in case writing the new keystore fails and
// leaves the keystore corrupted, but also include a --no-backup type of option
final
class SecretKeyGen
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final FileWrapper ksFile;
    private final String ksType;
    private final String alias;
    private final int keyLen;
    private final String algorithm;

    private
    SecretKeyGen( Configurator c )
    {
        super( c );

        this.ksFile = new FileWrapper( inputs.notNull( c.ksFile, "ksFile" ) );
        this.ksType = inputs.notNull( c.ksType, "ksType" );
        this.alias = inputs.notNull( c.alias, "alias" ).toLowerCase();
        this.keyLen = inputs.positiveI( c.keyLen, "keyLen" );
        this.algorithm = inputs.notNull( c.algorithm, "algorithm" );
    }

    private
    char[]
    getPassword( String fmt,
                 Object... args )
    {
        Console c = System.console();

        state.isFalse( 
            c == null, "No console available (Can't prompt for pass)" );

        return c.readPassword( fmt, args );
    }
        
    private
    KeyStore
    loadKeyStore( char[] keyStorePass )
        throws Exception
    {
        if ( ksFile.exists() )
        {
            KeyStore res = KeyStore.getInstance( ksType );

            FileInputStream fis = new FileInputStream( ksFile.getFile() );
            try
            {
                res.load( fis, keyStorePass );
                return res;
            }
            finally { fis.close(); }
        }
        else 
        {
            KeyStore res = KeyStore.getInstance( ksType );
            res.load( null );

            return res;
        }
    }

    private
    void
    checkAliasAbsent( KeyStore ks )
        throws Exception
    {
        Enumeration< String > en = ks.aliases();
        
        while ( en.hasMoreElements() )
        {
            inputs.isFalse( 
                en.nextElement().equals( alias ),
                "Key store already has a key for alias", alias );
        }
    }

    private
    SecretKey
    generateSecretKey()
        throws Exception
    {
        SecureRandom r = new SecureRandom();
        KeyGenerator kg = KeyGenerator.getInstance( algorithm );
        kg.init( keyLen );

        return kg.generateKey();
    }

    private
    void
    addKey( KeyStore ks,
            SecretKey key,
            char[] keyPass )
        throws Exception
    {
        KeyStore.PasswordProtection pp = 
            new KeyStore.PasswordProtection( keyPass );

        KeyStore.SecretKeyEntry e = new KeyStore.SecretKeyEntry( key );

        ks.setEntry( alias, e, pp );
    }

    private
    void
    saveKeyStore( KeyStore ks,
                  char[] keyStorePass )
        throws Exception
    {
        FileOutputStream fos = new FileOutputStream( ksFile.getFile() );
        try { ks.store( fos, keyStorePass ); } finally { fos.close(); }
    }

    public
    int
    execute()
        throws Exception
    {
        char[] keyStorePass = getPassword( "Keystore password: " );
        KeyStore ks = loadKeyStore( keyStorePass );

        checkAliasAbsent( ks );
        
        SecretKey key = generateSecretKey();
        char[] keyPass = getPassword( "Password for key %1$s: ", alias );
        addKey( ks, key, keyPass );

        saveKeyStore( ks, keyStorePass );

        return 0;
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private String ksFile;
        private String ksType = CryptoConstants.KEY_STORE_TYPE_JCEKS;
        private String alias;
        private int keyLen;
        private String algorithm;

        @Argument
        private
        void
        setKeyStoreFile( String ksFile )
        {
            this.ksFile = ksFile;
        }

        @Argument
        private void setKeyStoreType( String ksType ) { this.ksType = ksType; }

        @Argument
        private void setKeyAlias( String alias ) { this.alias = alias; }

        @Argument
        private void setKeyLength( int keyLen ) { this.keyLen = keyLen; }

        @Argument
        private
        void
        setAlgorithm( String algorithm )
        {
            this.algorithm = algorithm;
        }
    }
}
