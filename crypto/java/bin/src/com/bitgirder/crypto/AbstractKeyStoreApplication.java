package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.FileWrapper;

import com.bitgirder.application.ApplicationProcess;

import java.io.FileOutputStream;
import java.io.FileInputStream;
import java.io.Console;

import java.security.KeyStore;
import java.security.KeyStoreException;

public
abstract
class AbstractKeyStoreApplication
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final FileWrapper ksFile;
    private final String ksType;

    protected
    AbstractKeyStoreApplication( Configurator c )
    {
        super( inputs.notNull( c, "c" ) );

        this.ksFile = inputs.notNull( c.ksFile, "ksFile" );
        this.ksType = inputs.notNull( c.ksType, "ksType" );
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

    protected
    final
    static
    class KeyStoreLoad
    {
        private final KeyStore ks;
        private final char[] ksPass;

        private
        KeyStoreLoad( KeyStore ks,
                      char[] ksPass )
        {
            this.ks = ks;
            this.ksPass = ksPass;
        }

        public final KeyStore getKeyStore() { return ks; }
        public final char[] getKeyStorePass() { return ksPass; }

        // returns this so callers can inline clearing with assignment
        public
        final
        KeyStoreLoad
        clearKeyStorePassword()
        {
            for ( int i = 0, e = ksPass.length; i < e; ++i ) ksPass[ i ] = 0;

            return this;
        }
    }
 
    // loads ks from a file if file exists; otherwise creates a new keystore,
    // presumably because the caller will be storing it in the file's location
    // later
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

    protected
    final
    KeyStoreLoad
    loadKeyStore( boolean expctFile )
        throws Exception
    {
        if ( expctFile ) ksFile.assertExists();

        char[] pass = getPassword( "Keystore password: " );
        KeyStore ks = loadKeyStore( pass );

        return new KeyStoreLoad( ks, pass );
    }

    protected
    static
    abstract
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private FileWrapper ksFile;
        private String ksType = CryptoConstants.KEY_STORE_TYPE_JCEKS;
        private String alias;

        @Argument
        private
        void
        setKeyStoreFile( FileWrapper ksFile )
        {
            this.ksFile = ksFile;
        }

        @Argument
        private void setKeyStoreType( String ksType ) { this.ksType = ksType; }

        @Argument
        private void setKeyAlias( String alias ) { this.alias = alias; }
    }
}
