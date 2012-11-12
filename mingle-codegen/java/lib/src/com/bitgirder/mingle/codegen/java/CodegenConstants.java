package com.bitgirder.mingle.codegen.java;

final
class CodegenConstants
{
    final static JvTypeName JV_TYPE_BIND_IMPLEMENTATION =
        new JvTypeName( "AbstractBindImplementation" );

    final static JvQname JV_QNAME_JLANG_EXCEPTION =
        JvQname.create( "java.lang", "Exception" );

    final static JvQname JV_QNAME_JLANG_THROWABLE =
        JvQname.create( "java.lang", "Throwable" );

    final static JvQname JV_QNAME_VOID = JvQname.create( "java.lang", "Void" );
    final static JvQname JV_QNAME_LONG = JvQname.create( "java.lang", "Long" );

    final static JvQname JV_QNAME_INT = 
        JvQname.create( "java.lang", "Integer" );

    final static JvQname JV_QNAME_DOUBLE = 
        JvQname.create( "java.lang", "Double" );

    final static JvQname JV_QNAME_FLOAT = 
        JvQname.create( "java.lang", "Float" );

    final static JvQname JV_QNAME_OBJECT = 
        JvQname.create( "java.lang", "Object" );

    final static JvQname JV_QNAME_PROC_ACTIVITY =
        JvQname.create( "com.bitgirder.process", "ProcessActivity" );

    final static JvType JV_TYPE_PROC_ACTIVITY_CTX =
        JvTypeExpression.dotTypeOf(
            JV_QNAME_PROC_ACTIVITY, new JvTypeName( "Context" ) );

    final static JvQname JV_QNAME_PROCESS_OPERATION =
        JvQname.create( "com.bitgirder.process", "ProcessOperation" );

    final static JvQname JV_QNAME_MINGLE_VALUE =
        JvQname.create( "com.bitgirder.mingle.model", "MingleValue" );

    final static JvQname JV_QNAME_JLIST = JvQname.create( "java.util", "List" );

    final static JvQname JV_QNAME_MINGLE_BINDERS =
        JvQname.create( "com.bitgirder.mingle.bind", "MingleBinders" );

    final static JvType JV_TYPE_BINDERS_INITIALIZER =   
        JvTypeExpression.dotTypeOf(
            JV_QNAME_MINGLE_BINDERS, new JvTypeName( "Initializer" ) );

    final static JvId JV_ID_VALIDATE_FIELD_VALUE = 
        new JvId( "validateFieldValue" );

    final static JvQname JV_QNAME_MG_TIMESTAMP =
        JvQname.create( "com.bitgirder.mingle.model", "MingleTimestamp" );

    final static JvType JV_QNAME_TYPE_DEF_LOOKUP =
        JvQname.create( "com.bitgirder.mingle.model", "TypeDefinitionLookup" );
       
    final static JvQname JV_QNAME_BYTE_BUFFER =
        JvQname.create( "java.nio", "ByteBuffer" );

    final static JvId JV_ID_NULL = new JvId( "null" );

    final static JvTypeName JV_TYPE_WILDCARD = new JvTypeName( "?" );

    final static JvQname JV_QNAME_MAP = JvQname.create( "java.util", "Map" );

    final static JvQname JV_QNAME_CLASS = 
        JvQname.create( "java.lang", "Class" );

    final static JvQname JV_QNAME_BG_LANG = 
        JvQname.create( "com.bitgirder.lang", "Lang" );

    final static JvId JV_ID_CAST_UNCHECKED = new JvId( "castUnchecked" );

    final static JvQname JV_QNAME_INPUTS =
        JvQname.create( "com.bitgirder.validation", "Inputs" );

    final static JvId JV_ID_NOT_NULL = new JvId( "notNull" );

    final static JvType JV_TYPE_IMMUTABLE_LIST_BUILDER =
        JvTypeExpression.dotTypeOf( 
            JV_QNAME_BG_LANG,
            new JvTypeName( "ImmutableListBuilder" )
        );
    
    final static JvId JV_ID_BUILD = new JvId( "build" );
    final static JvId JV_ID_ADD = new JvId( "add" );
    final static JvId JV_ID_VALUE_OF = new JvId( "valueOf" );

    final static JvQname JV_QNAME_FIELD_DEF =
        JvQname.create( "com.bitgirder.mingle.model", "FieldDefinition" );

    final static JvQname JV_QNAME_MG_ENUM =
        JvQname.create( "com.bitgirder.mingle.model", "MingleEnum" );

    final static JvQname JV_QNAME_MG_IDENTIFIER =
        JvQname.create( "com.bitgirder.mingle.model", "MingleIdentifier" );

    final static JvQname JV_QNAME_MG_NAMESPACE =
        JvQname.create( "com.bitgirder.mingle.model", "MingleNamespace" );

    final static JvQname JV_QNAME_MG_TYPE_REF =
        JvQname.create( "com.bitgirder.mingle.model", "MingleTypeReference" );

    final static JvQname JV_QNAME_ATOMIC_TYPE_REF =
        JvQname.create( "com.bitgirder.mingle.model", "AtomicTypeReference" );

    final static JvQname JV_QNAME_MG_QNAME =
        JvQname.create( "com.bitgirder.mingle.model", "QualifiedTypeName" );

    final static JvQname JV_QNAME_BOUND_SERVICE =
        JvQname.create( "com.bitgirder.mingle.bind", "BoundService" );

    final static JvQname JV_QNAME_SYM_MAP_BUILDER =
        JvQname.create( 
            "com.bitgirder.mingle.model", "MingleSymbolMapBuilder" );

    final static JvQname JV_QNAME_SYM_MAP =
        JvQname.create( "com.bitgirder.mingle.model", "MingleSymbolMap" );

    final static JvQname JV_QNAME_MINGLE_BINDER =
        JvQname.create( "com.bitgirder.mingle.bind", "MingleBinder" );

    final static JvType JV_TYPE_BINDER_BUILDER =
        JvTypeExpression.dotTypeOf(
            JV_QNAME_MINGLE_BINDER, new JvTypeName( "Builder" ) );

    final static JvQname JV_QNAME_MG_RUNTIME =    
        JvQname.create( "com.bitgirder.mingle.runtime", "MingleRuntime" );

    final static JvQname JV_QNAME_JSTRING = 
        JvQname.create( "java.lang", "String" );

    final static JvQname JV_QNAME_OBJ_PATH =
        JvQname.create( "com.bitgirder.lang.path", "ObjectPath" );

    final static JvId JV_ID_GET_ROOT = new JvId( "getRoot" );

    final static JvType JV_TYPE_OBJ_PATH_STRING =
        JvTypeExpression.withParams( JV_QNAME_OBJ_PATH, JV_QNAME_JSTRING ); 
    
    final static JvId JV_ID_CREATE = new JvId( "create" );
    
    final static JvId JV_ID_AS_JAVA_VALUE = new JvId( "asJavaValue" );
}
