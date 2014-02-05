require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/io'
include BitGirder::Io

require 'erb'

class AccDef < BitGirderClass

    bg_attr :mg_type

    bg_attr :mg_type_var, required: false

    bg_attr :uses_mg_type_var, default: true

    bg_attr :jv_type, required: false

    bg_attr :jv_getter_type_name,
            required: false,
            description: "type name for jv getter"

    bg_attr :jv_res_exp, 
            required: false,
            description: "sprintf with %s arg for the starting val"

    bg_attr :jv_null_val,
            default: "null",
            description: "return this when mingle val is null"

    private
    def impl_initialize
        
        super

        unless @mg_type_var || ( ! @uses_mg_type_var )
 
            nm_base = 
                @mg_type.sub( /^Mingle/, "" ).
                gsub( /([^[:upper:]])([[:upper:]])/ ) { |m| "#$1_#$2" }.
                upcase

            @mg_type_var = "Mingle.TYPE_#{nm_base}"
        end
    end

    private
    def uc0( s )
        "#{ s[ 0 ].upcase }#{ s[ 1 .. -1 ] }"
    end

    private
    def getter_name( base, pref )
        "#{pref}#{ uc0( base ) }"
    end

    public
    def mg_getter_name( pref )
        getter_name( @mg_type, pref )
    end

    public
    def jv_getter_name( pref )
        getter_name( @jv_getter_type_name || @jv_type, pref )
    end
end

class Generator < AbstractApplication
    
    private
    def create_acc_defs
        
        [ 
            AccDef.new(
                mg_type: "MingleBoolean",
                jv_type: "boolean",
                jv_res_exp: "%s.booleanValue()",
                jv_null_val: "false",
            ),

            AccDef.new(
                mg_type: "MingleInt32",
                jv_type: "int",
                jv_res_exp: "%s.intValue()",
                jv_null_val: "0",
            ),

            AccDef.new(
                mg_type: "MingleInt64",
                jv_type: "long",
                jv_res_exp: "%s.longValue()",
                jv_null_val: "0",
            ),

            AccDef.new(
                mg_type: "MingleUint32",
                jv_type: "int",
                jv_getter_type_name: "uint",
                jv_res_exp: "%s.intValue()",
                jv_null_val: "0",
            ),

            AccDef.new(
                mg_type: "MingleUint64",
                jv_type: "long",
                jv_getter_type_name: "ulong",
                jv_res_exp: "%s.longValue()",
                jv_null_val: "0",
            ),

            AccDef.new(
                mg_type: "MingleFloat32",
                jv_type: "float",
                jv_res_exp: "%s.floatValue()",
                jv_null_val: "0.0f",
            ),

            AccDef.new(
                mg_type: "MingleFloat64",
                jv_type: "double",
                jv_res_exp: "%s.doubleValue()",
                jv_null_val: "0.0d",
            ),

            AccDef.new(
                mg_type: "MingleString",
                jv_type: "String",
                jv_res_exp: "%s.toString()",
            ),

            AccDef.new(
                mg_type: "MingleBuffer",
                jv_type: "byte[]",
                jv_res_exp: "%s.array()",
                jv_getter_type_name: "byteArray",
            ),

            AccDef.new( mg_type: "MingleTimestamp" ),
            AccDef.new( mg_type: "MingleSymbolMap" ),

            AccDef.new( 
                mg_type: "MingleList",
                mg_type_var: "Mingle.TYPE_VALUE_LIST",
                jv_type: "MingleListAccessor",
                jv_getter_type_name: "listAccessor",
                jv_res_exp: "MingleListAccessor.forList( %s, %s )",
            ),

            AccDef.new( mg_type: "MingleValue" ),

            AccDef.new(
                mg_type: "MingleStruct",
                uses_mg_type_var: false,
                jv_type: "MingleStructAccessor",
                jv_getter_type_name: "structAccessor",
                jv_res_exp: "MingleStructAccessor.forStruct( %s, %s )",
            ),
        ]
    end

    # prefix output with line numbers, helpful in debugging compilation errors
    # in generated code (which of course should only happen when working on this
    # script)
    private
    def with_line_numbers( src )
        
        lines = src.split( /\n/ )

        res = Array.new( lines.size ) do |i|
            line = sprintf( "/* %5i */ ", i + 1 )
            line << lines[ i ]
        end

        res.join( "\n" )
    end

    public
    def autogen_header
        
        "// Autogenerated on #{Time.now}"
    end

    private
    def generate( nm_sym, acc_defs )

        tmpl = ERB.new( Kernel.const_get( nm_sym ) )

        gen_dir = has_key( ENV, "CODEGEN_GEN_SRC_DIR" )
        rel_name = "com/bitgirder/mingle/#{nm_sym}.java"
        out_file = "#{gen_dir}/#{rel_name}"

        src = tmpl.result( binding )
        src = with_line_numbers( src )

        File.open( ensure_parent( out_file ), "w" ) { |io| io.print src }
    end

    private
    def impl_run

        acc_defs = create_acc_defs

        [ :MingleSymbolMapAccessor, :MingleListAccessor ].each do |nm_sym|
            generate( nm_sym, acc_defs )
        end
    end

end

MingleSymbolMapAccessor = <<END
<%= autogen_header %>

package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import java.util.List;

public
class MingleSymbolMapAccessor
extends MingleValueAccessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleSymbolMap map;
    private final ObjectPath< MingleIdentifier > path;

    MingleSymbolMapAccessor( MingleSymbolMap map,
                             ObjectPath< MingleIdentifier > path )
    {
        this.map = map;
        this.path = path;
    }

    public final MingleSymbolMap getMap() { return map; }
    public final ObjectPath< MingleIdentifier > getPath() { return path; }

    <% acc_defs.each do |acc_def| %>

    <%
        mg_impl_getter = acc_def.mg_getter_name( "access" )

        jv_impl_getter = if acc_def.jv_type
            acc_def.jv_getter_name( "access" )
        else
            nil
        end
    %>

    // May return null, but never MingleNull
    private
    <%= acc_def.mg_type %>
    <%= mg_impl_getter %>( MingleIdentifier fld,
                           boolean expct )
    {
        MingleValue res = accessValue(
            map.get( fld ),
            path.descend( fld ),
            <%= acc_def.uses_mg_type_var ? acc_def.mg_type_var : "null" %>,
            <%= acc_def.mg_type %>.class
        );

        if ( res instanceof MingleNull ) {

            if ( expct ) {
                List< MingleIdentifier > flds = Lang.singletonList( fld );
                throw new MingleMissingFieldsException( flds, path );
            }

            return null;
        }

        return (<%= acc_def.mg_type %>) res;
    }

    <% if jv_impl_getter %>

    private
    <%= acc_def.jv_type %>
    <%= jv_impl_getter %>( MingleIdentifier fld,
                           boolean expct )
    {
        MingleValue res = <%= mg_impl_getter %>( fld, expct );
        if ( res == null ) return <%= acc_def.jv_null_val %>;

        <%= acc_def.mg_type %> typedRes = (<%= acc_def.mg_type %>) res;
        return <%= 
            sprintf( acc_def.jv_res_exp, "typedRes", "path.descend( fld )" ) 
        %>;
    }

    <% end %> <%# if jv_impl_getter %>

    <% %w{ get expect }.each do |get_type| %>
    
    <%
        mg_getter = acc_def.mg_getter_name( get_type )
        expct = ( get_type == "expect" ).to_s # "true"|"false"
    %>

    public
    final
    <%= acc_def.mg_type %>
    <%= mg_getter %>( MingleIdentifier fld )
    {
        inputs.notNull( fld, "fld" );
        return <%= mg_impl_getter %>( fld, <%= expct %> );
    }

    public
    final
    <%= acc_def.mg_type %>
    <%= mg_getter %>( CharSequence fld )
    {
        inputs.notNull( fld, "fld" );
        MingleIdentifier fldId = MingleIdentifier.create( fld );
        return <%= mg_impl_getter %>( fldId, <%= expct %> );
    }

    <% if jv_impl_getter %>

    <%
        jv_getter = acc_def.jv_getter_name( get_type )
    %>

    public
    final
    <%= acc_def.jv_type %>
    <%= jv_getter %>( CharSequence fld )
    {
        inputs.notNull( fld, "fld" );
        MingleIdentifier fldId = MingleIdentifier.create( fld );
        return <%= jv_impl_getter %>( fldId, <%= expct %> );
    }

    public
    final
    <%= acc_def.jv_type %>
    <%= jv_getter %>( MingleIdentifier fld )
    {
        inputs.notNull( fld, "fld" );
        return <%= jv_impl_getter %>( fld, <%= expct %> );
    }

    <% end %> <%# expcect|get block %>

    <% end %> <%# if %>

    <% end %> <%# each %>

    private
    MingleStruct
    implAccessMingleStruct( MingleIdentifier fld,
                            boolean expct )
    {
        MingleValue mv = accessMingleValue( fld, expct );
        if ( mv == null ) return null;

        if ( mv instanceof MingleStruct ) return (MingleStruct) mv;

        throw new MingleValueCastException(
            "expected a struct but found: " + Mingle.inferredTypeOf( mv ),
            path.descend( fld )
        );
    }

    public
    static
    MingleSymbolMapAccessor
    forMap( MingleSymbolMap map,
            ObjectPath< MingleIdentifier > path )
    {
        return new MingleSymbolMapAccessor(
            inputs.notNull( map, "map" ),
            inputs.notNull( path, "path" )
        );
    }

    public
    static
    MingleSymbolMapAccessor
    forMap( MingleSymbolMap map )
    {
        return forMap( map, ObjectPath.< MingleIdentifier >getRoot() );
    }
}

END

MingleListAccessor = <<END
<%= autogen_header %>
package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.MutableListPath;

import java.util.Iterator;

public
final
class MingleListAccessor
extends MingleValueAccessor
implements Iterable< MingleValue >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleList list;

    // root path for this list -- should be used to create a list path as needed
    private final ObjectPath< MingleIdentifier > root;

    private
    MingleListAccessor( MingleList list,
                        ObjectPath< MingleIdentifier > root )
    {
        this.list = list;
        this.root = root;
    }

    public Iterator< MingleValue > iterator() { return list.iterator(); }

    public
    final
    class Traversal
    implements Iterator< MingleValue >
    {
        private final MutableListPath< MingleIdentifier > mlp = 
            MingleListAccessor.this.root.startMutableList();

        private final Iterator< MingleValue > it = 
            MingleListAccessor.this.iterator();

        private Traversal() {}

        public boolean hasNext() { return it.hasNext(); }

        public 
        MingleValue 
        next() 
        { 
            MingleValue res = it.next(); 
            mlp.increment();

            return res;
        }

        public void remove() { it.remove(); }

        <% acc_defs.each do |acc_def| %>

        <%
            next_getter = acc_def.mg_getter_name( "next" )
        %>

        public
        <%= acc_def.mg_type %>
        <%= next_getter %>()
        {
            MingleValue res = accessValue( 
                next(),
                mlp,
                <%= acc_def.uses_mg_type_var ? acc_def.mg_type_var : "null" %>,
                <%= acc_def.mg_type %>.class
            );

            if ( res instanceof MingleNull ) return null;
            return (<%= acc_def.mg_type %>) res;
        }

        <% if acc_def.jv_type %>

        public
        <%= acc_def.jv_type %>
        <%= acc_def.jv_getter_name( "next" ) %>()
        {
            <%= acc_def.mg_type %> res = <%= next_getter %>();

            if ( res == null ) return <%= acc_def.jv_null_val %>;

            return <%= 
                sprintf( acc_def.jv_res_exp, "res", "mlp.createCopy()" ) 
            %>;
        }
        
        <% end %> <%# if jv_type %>

        <% end %> <%# acc_defs block %>
    }

    public Traversal traversal() { return new Traversal(); }

    public
    static
    MingleListAccessor
    forList( MingleList list,
             ObjectPath< MingleIdentifier > root )
    {
        return new MingleListAccessor(
            inputs.notNull( list, "list" ),
            inputs.notNull( root, "root" )
        );
    }

    public
    static
    MingleListAccessor
    forList( MingleList list )
    {
        return forList( list, ObjectPath.< MingleIdentifier >getRoot() );
    }
}

END

BitGirderCliApplication.run( Generator )
