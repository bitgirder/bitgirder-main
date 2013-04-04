goog.provide( "gird.log" );

goog.require( "goog.string.format" );
goog.require( "goog.json" );

goog.scope( function() {

"use strict";

var Logger = function() {};

function sendConsole( meth, argArr ) {
    if ( typeof( console ) !== 'undefined' ) {
        console[ meth ].apply( console, argArr );
    }
}

Logger.prototype.log = function( msg ) { sendConsole( "log", [ msg ] ); };

Logger.prototype.logf = function( tmpl, args ) {
    this.log( goog.string.format.apply( null, arguments ) );
};

Logger.prototype.dumpObject = function( obj ) { 
    return goog.json.serialize( obj ); 
};

gird.log.getLogger = function() { return new Logger(); };

});
