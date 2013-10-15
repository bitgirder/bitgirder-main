goog.provide( "gird.testing" );

goog.require( "gird.log" );
goog.require( "gird.core" );

goog.require( "goog.events" );
goog.require( "goog.events.EventTarget" );
goog.require( "goog.ui.IdGenerator" );

goog.scope( function() {

var $log = gird.log.getLogger();
var $inputs = gird.core.Inputs;
var $mod = gird.testing;
var $idGen = goog.ui.IdGenerator.getInstance();

function mkEvName( nm ) { return "gird:testing." + nm; }

$mod.EventType = { ALL: [] };

function mkEvType( nm ) { 

    var ev = "gird:testing." + nm;

    $mod.EventType[ nm ] = ev;
    $mod.EventType.ALL.push( ev );
}

mkEvType( "ADD_TEST" );
mkEvType( "START_RUN" );
mkEvType( "COMPLETE_RUN" );
mkEvType( "START_TEST" );
mkEvType( "COMPLETE_TEST" );

$mod.TestStateError = function( msg ) { this.message = msg; };
goog.inherits( $mod.TestStateError, Error );

$mod.TestStateError.toString = function() { return this.msg; };

// Make class private; we expose the instance via gird.testing.TestRun
function TestRun() {

    goog.events.EventTarget.call( this );

    this.tests = [];
    this.active = 0;
};

goog.inherits( TestRun, goog.events.EventTarget );

// Single global instance.
$mod.TestRun = new TestRun();

// Impl is simple for now; later this may be more elaborate if we have
// sequencing in which code that uses a test run might execute before the test
// run itself is present. In that case, this method might queue the 'cb'
// parameter until the test run is ready to be used.
$mod.withTestRun = function( cb ) { cb( $mod.TestRun ); };

TestRun.prototype.addTest = function( name, fn ) {
 
    $inputs.notNull( name, "name" );
    $inputs.notNull( fn, "fn" );

    if ( this.running ) {
        throw new $mod.TestStateError( 
            "call to addTest() when tests are running" );
    }

    var id = "test:" + $idGen.getNextUniqueId();
    var test = { id: id, name: name, fn: fn };

    this.tests.push( test );
    trDispatch( this, { type: $mod.EventType.ADD_TEST, test: test } );

    ++this.active;
};

function PackageTestContext( pkg, tr ) { 
    
    this.pkg = pkg;
    this.tr = tr; 
}

PackageTestContext.prototype.addTest = function() {

    arguments[ 0 ] = this.pkg + "/" + arguments[ 0 ];
    this.tr.addTest.apply( this.tr, arguments );
};

TestRun.prototype.forPackage = function( pkg, optCb ) {
    
    var ptc = new PackageTestContext( pkg, this );

    if ( optCb ) { optCb( ptc ); }

    return ptc;
};

$mod.forPackage = function( pkg, cb ) {
    $mod.withTestRun( function( tr ) { tr.forPackage( pkg, cb ); } );
};

function trDispatch( tr, ev ) { tr.dispatchEvent( ev ); }

function trCompleteRun( tr ) {
    trDispatch( tr, { type: $mod.EventType.COMPLETE_RUN } );
}

function TestContext( tr, test ) {

    this.tr = tr;
    this.test = test;
}

function createTestContext( tr, test ) { return new TestContext( tr, test ); }

function tcComplete( tc, err ) {

    if ( ! tc.done ) {

        tc.done = true;
    
        trDispatch( tc.tr, { 
            type: $mod.EventType.COMPLETE_TEST, 
            test: tc.test, 
            err: err
        });

        if ( --tc.tr.active == 0 ) { trCompleteRun( tc.tr ); }
    }
}

TestContext.prototype.fail = function( err ) { tcComplete( this, err ); }
TestContext.prototype.ok = function() { tcComplete( this, null ); }

function trBeginTest( tr, test ) {

    trDispatch( tr, { type: $mod.EventType.START_TEST, test: test } );

    var tc = createTestContext( tr, test );

    try { test.fn( tc ); } catch ( err ) { return tc.fail( err ); }
    if ( test.fn.length === 0 ) { tc.ok(); }
}

TestRun.prototype.begin = function() {

    trDispatch( this, { type: $mod.EventType.START_RUN } );

    if ( goog.array.isEmpty( this.tests ) ) { return trCompleteRun( this ); }

    goog.array.forEach( this.tests, function( test ) {
        trBeginTest( this, test );
    }, this );
};

});
