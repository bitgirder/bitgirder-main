package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Enumeration;
import java.util.SortedSet;

import java.security.KeyStore;

final
class KeyStoreUtil
extends AbstractKeyStoreApplication
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private 
    static 
    enum Operation 
    {
        LIST_KEY_ALIASES;
    }

    private final Operation op;

    private
    KeyStoreUtil( Configurator c )
    {
        super( c );

        this.op = inputs.notNull( c.op, "op" );
    }

    private
    void
    listKeyAliases( KeyStore ks )
        throws Exception
    {
        SortedSet< String > aliases = Lang.newSortedSet();

        for ( Enumeration< String > en = ks.aliases(); en.hasMoreElements(); )
        {
            aliases.add( en.nextElement() );
        }

        for ( String alias : aliases ) System.out.println( alias );
    }

    public
    int
    execute()
        throws Exception
    {
        KeyStoreLoad ksl = loadKeyStore( true ).clearKeyStorePassword();

        switch ( op )
        {
            case LIST_KEY_ALIASES: listKeyAliases( ksl.getKeyStore() ); break;
        }

        return 0;
    }

    private
    final
    static
    class Configurator
    extends AbstractKeyStoreApplication.Configurator
    {
        private Operation op;

        @Argument private void setOperation( Operation op ) { this.op = op; }
    }
}
