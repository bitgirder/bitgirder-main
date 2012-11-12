package com.bitgirder.aws;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.Charsets;

import com.bitgirder.application.ApplicationProcess;

import com.bitgirder.crypto.CryptoUtils;

import java.io.Console;
import java.io.InputStream;
import java.io.FileInputStream;
import java.io.FileOutputStream;

import java.security.KeyStore;

import javax.crypto.spec.SecretKeySpec;

final
class AwsKeyStoreUtil
extends ApplicationProcess
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static enum Operation { ADD_SECRET_KEY; }

    private final FileWrapper keyFile;
    private final String keyAlias;
    private final Operation op;

    // hardcoded at this point, may be configurable at some point
    private final String keyStoreType = "jceks"; 
    private final String keyAlg = "HmacSHA1";

    private
    AwsKeyStoreUtil( Configurator c )
    {
        super( c );

        this.keyFile = inputs.notNull( c.keyFile, "keyFile" );
        this.keyAlias = inputs.notNull( c.keyAlias, "keyAlias" );
        this.op = inputs.notNull( c.op, "op" );
    }

    private
    KeyStore
    loadKeyStore( char[] pass )
        throws Exception
    {
        KeyStore res = KeyStore.getInstance( keyStoreType );

        InputStream is = 
            keyFile.exists() ? new FileInputStream( keyFile.getFile() ) : null;
        
        try { res.load( is, pass ); } finally { if ( is != null ) is.close(); }

        return res;
    }

    // The current implementation uses a String when converting from char[] to
    // UTF-8 byte[], but we can avoid that later if we get more paranoid about
    // wanting to be able to zero out any memory containing the char[] form of
    // the secret key. Since this program is not intended to be long-lived nor
    // used with other untrusted code, using a String seems like a reasonable
    // risk at this point.
    private
    SecretKeySpec
    getSecretKeySpec()
        throws Exception
    {
        char[] keyChars = CryptoUtils.readPassword( "AWS Secret key: " );
        String keyStr = new String( keyChars );
        byte[] keyBytes = keyStr.getBytes( Charsets.UTF_8.charset() );

        return new SecretKeySpec( keyBytes, keyAlg );
    }

    private
    void
    saveKeyStore( KeyStore ks,
                  char[] ksPass )
        throws Exception
    {
        FileOutputStream os = new FileOutputStream( keyFile.getFile() );
        try { ks.store( os, ksPass ); } finally { os.close(); }
    }

    private
    char[]
    readKeyPass()
    {
        char[] keyPass = 
            CryptoUtils.readPassword( "AWS Secret key password: " );

        char[] keyPassConfirm =
            CryptoUtils.readPassword( "Confirm AWS Secret key password: " );

        inputs.isTrue(
            new String( keyPass ).equals( new String( keyPassConfirm ) ),
            "Password and confirmation password do not match" );
        
        return keyPass;
    }

    private
    void
    doAddSecretKey()
        throws Exception
    {
        SecretKeySpec keySpec = getSecretKeySpec();
        KeyStore.SecretKeyEntry ske = new KeyStore.SecretKeyEntry( keySpec );

        char[] keyPass = readKeyPass();

        KeyStore.PasswordProtection param =
            new KeyStore.PasswordProtection( keyPass );

        char[] ksPass = CryptoUtils.readPassword( "KeyStore password: " );
        KeyStore ks = loadKeyStore( ksPass );
        ks.setEntry( keyAlias, ske, param );

        saveKeyStore( ks, ksPass );
    }

    protected
    void
    startImpl()
        throws Exception
    {
        switch ( op )
        {
            case ADD_SECRET_KEY: doAddSecretKey(); break;
        }

        exit();
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private FileWrapper keyFile;
        private String keyAlias;
        private Operation op;

        @Argument
        private
        void
        setKeyFile( FileWrapper keyFile )
        {
            this.keyFile = keyFile;
        }

        @Argument 
        private 
        void 
        setKeyAlias( String keyAlias ) 
        { 
            this.keyAlias = keyAlias; 
        }

        @Argument private void setOperation( Operation op ) { this.op = op; }
    }
}
