package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;

import java.util.regex.Pattern;
import java.util.regex.Matcher;

final
class JvQname
implements JvExpression,
           JvType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // currently tailored to our particular needs: does not (yet) allow all
    // legal java qnames, but is such that all qnames allowed are valid java
    // qnames
    private final static String PKG_NAME_PART = "[a-z][a-z_\\d]*";
    private final static Pattern PARSE_PAT =
        PatternHelper.compile(
            "^(" + PKG_NAME_PART + "(?:\\." + PKG_NAME_PART + ")*)" +
            "\\." +
            "([A-Z].*)$"
        );

    final JvPackage pkg;
    final JvTypeName nm;

    JvQname( JvPackage pkg,
             JvTypeName nm )
    {
        this.pkg = state.notNull( pkg, "pkg" );
        this.nm = state.notNull( nm, "nm" );
    }

    public void validate() {}

    @Override public String toString() { return pkg + "." + nm; }

    @Override public int hashCode() { return pkg.hashCode() ^ nm.hashCode(); }

    @Override
    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        else if ( o instanceof JvQname )
        {
            JvQname q = (JvQname) o;
            return q.pkg.equals( pkg ) && q.nm.equals( nm );
        }
        else return false;
    }

    static
    JvQname
    forUnit( JvCompilationUnit u )
    {
        return create( u.pkg, u.decl.name );
    }

    static
    JvQname
    create( CharSequence pkg,
            CharSequence nm )
    {
        inputs.notNull( pkg, "pkg" );
        inputs.notNull( nm, "nm" );

        return new JvQname( new JvPackage( pkg ), new JvTypeName( nm ) );
    }

    static
    JvQname
    parse( CharSequence str )
    {
        Matcher m = state.matches( str, "str", PARSE_PAT );

        return create( m.group( 1 ), m.group( 2 ) );
    }
}
