goog.provide( "gird.testing.assert" );

goog.require( "gird.core" );
goog.require( "gird.log" );

goog.scope( function() {

var $mod = gird.testing.assert;
var $log = gird.log.getLogger();
var $core = gird.core;

$mod.fail = function( msg ) { throw new Error( msg ); }

$mod.failf = function( tmpl, args ) {
    $mod.fail( $core.sprintf.apply( null, arguments ) );
};

// Passes with any truthy value; if we want to explicitly test for 'true' we can
// add other ways to do that, including calling this method as:
//
//  isTrue( val === true, ... )
//
$mod.isTrue = function( val, msg ) {
    
    if ( val ) { return; }
    $mod.fail( msg );
};

$mod.isTruef = function( val, tmpl, args ) {
    
    if ( arguments[ 0 ] ) { return; }

    $mod.failf.apply( null, Array.prototype.slice.call( arguments, 1 ) );
};

$mod.equalsf = function( expct, act, tmpl, args ) {

    $mod.isTruef.apply( null,
        [ expct === act ].concat( Array.prototype.slice.call( arguments, 2 ) )
    );
};

$mod.equals = function( expct, act ) {
    $mod.equalsf( expct, act, "expected %s but got %s", expct, act );
};

function deepEqArray( eq ) {
 
    $mod.isTruef( eq.ctx.isArrayFn( eq.act ),
        "expected array, got %s", typeof eq.act );
    
    $mod.isTrue(
        goog.array.equals( eq.expect, eq.act, function( a, b ) {
            deepEq( { expect: a, act: b, ctx: eq.ctx } );
            return true;
        }),
        "arrays are not equal"
    );
}

function deepEqString( eq ) {
    
    $mod.isTruef( goog.isString( eq.act ), "not a string: %s", eq.act );

    $mod.equals( eq.expect, eq.act );
}

function deepEq( eq ) {

    if ( eq.ctx.isArrayFn( eq.expect ) ) { return deepEqArray( eq ); }
    if ( goog.isString( eq.expect ) ) { return deepEqString( eq ); }

    $mod.failf( "unhandled expect value (typeof: %s)", typeof eq.expect );
}

$mod.deepEquals = function( expct, act ) {
 
    deepEq({
        expect: expct,
        act: act,
        ctx: { isArrayFn: goog.isArray }
    });
};

});
