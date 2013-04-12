goog.provide( "gird.core" );

goog.require( "goog.string.format" );

goog.scope( function() {

var $mod = gird.core;

$mod.sprintf = goog.string.format;

$mod.newErrorf = function( tmpl, args ) {
    return new Error( $mod.sprintf.apply( null, arguments ) );
}

$mod.Inputs = {};

function notNull( val, msgArgArr ) {

    if ( val == null ) {
        throw new Error( $mod.sprintf.apply( null, msgArgArr ) );
    }

    return val;
}

$mod.Inputs.notNull = function( val, nm ) {
    return notNull( val, [ "parameter '%s' is null or undefined", nm ] ); 
};

function hasKeyOrProperty( obj, key, objNm, errTmpl ) {

    $mod.Inputs.notNull( obj, "obj" );
    $mod.Inputs.notNull( key, "key" );
    $mod.Inputs.notNull( objNm, "objNm" );

    return notNull( obj[ key ], [ errTmpl, objNm, key ] );
}

$mod.Inputs.hasKey = function( obj, key, objNm ) {
    return hasKeyOrProperty( obj, key, objNm, 
        "parameter '%s' has no value for key '%s'" );
}

$mod.Inputs.hasProperty = function( obj, key, objNm ) {
    return hasKeyOrProperty( obj, key, objNm, 
        "%s has no value for property '%s'" );
};

});
