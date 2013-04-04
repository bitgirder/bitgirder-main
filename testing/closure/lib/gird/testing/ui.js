goog.provide( "gird.testing.ui" );

goog.require( "gird.testing" );
goog.require( "gird.log" );
goog.require( "gird.core" );

goog.require( "goog.events" );
goog.require( "goog.dom" );
goog.require( "goog.dom.classes" );
goog.require( "goog.structs" );

goog.scope( function() {

var $mod = gird.testing.ui;

var $log = gird.log.getLogger();
var $inputs = gird.core.Inputs;

var $evType = gird.testing.EventType;
var $dom = goog.dom;

function HtmlReporter( opts ) {

    this.element = $inputs.hasKey( opts, "element", "opts" );
    this.rowsById = {};
}

var $rpt = HtmlReporter.prototype;

var CLASS_TEST_STATUS = "test-status";

var TestStatus = {
    
    INITIALIZED: { cssClass: "initialized", text: "Initialized" },

    RUNNING: { cssClass: "running", text: "Running" },

    SUCCEEDED: { cssClass: "succeeded", text: "Succeeded" },

    FAILED: { cssClass: "failed", text: "Failed" }

};

var ALL_TEST_STATUS_CLASSES = goog.structs.map(
    goog.structs.getValues( TestStatus ),
    function( stat ) { return stat.cssClass }
);

$rpt.createTestRunTable = function() {
    
    this.tbl = $dom.createDom( 'table', "test-run-table",
        $dom.createDom( 'th', "test-run-table-header",
            $dom.createDom( 'td', '', "Test" ),
            $dom.createDom( 'td', '', "Status" )
        )
    );

    return this.tbl;
};

$rpt.setRunStatus = function( txt ) {
    $dom.setTextContent( this.statusSpan, txt );
};

$rpt.initDom = function() {

    $dom.removeChildren( this.element ); 

    $dom.appendChild( this.element,
        $dom.createDom( 'div', "test-run-root",
            $dom.createDom( 'div', "test-run-status",
                this.statusSpan = $dom.createDom( 'span', null, "" )
            ),
            $dom.createDom( 'div', "test-run-tests", this.createTestRunTable() )
        )
    );

    this.setRunStatus( "Loading..." );
}

$rpt.init = function( tr ) {

    goog.events.listen( tr, $evType.ALL, goog.bind( this.testEvent, this ) );
    this.initDom();
}

$rpt.expectTestRow = function( test ) {
    
    var row = this.rowsById[ test.id ];

    if ( row ) { return row; }
    throw gird.core.newErrorf( "No row for test with id %s", test.id );
}

$rpt.setTestStatus = function( test, stat ) {
 
    var row = this.expectTestRow( test );
    var elt = $dom.getElementByClass( CLASS_TEST_STATUS, row );
    goog.dom.classes.addRemove( elt, ALL_TEST_STATUS_CLASSES, stat.cssClass );
    $dom.setTextContent( elt, stat.text );
};

$rpt.addTest = function( test ) {
 
    var row = $dom.createDom( "tr", "", 
        $dom.createDom( "td", "test-name", test.name ),
        $dom.createDom( "td", CLASS_TEST_STATUS, "" )
    );

    // Must set the row by id before calling setTestStatus
    this.rowsById[ test.id ] = row;
    this.setTestStatus( test, TestStatus.INITIALIZED );

    $dom.appendChild( this.tbl, row );
}

$rpt.startRun = function() { this.setRunStatus( "Running..." ); };
$rpt.completeRun = function() { this.setRunStatus( "Complete" ); };

$rpt.startTest = function( test ) {
    this.setTestStatus( test, TestStatus.RUNNING );
};

$rpt.addError = function( test, err ) {
 
    $log.log( err );
    var stack = err.stack;

    if ( stack ) {
        $log.logf( "Error for %s: %s", test.name, stack );
    }
};

$rpt.completeTest = function( test, err ) {

    var stat = err ? TestStatus.FAILED : TestStatus.SUCCEEDED;
    this.setTestStatus( test, stat );

    if ( err ) { this.addError( test, err ); }
};

$rpt.testEvent = function( ev ) {
    
    switch ( ev.type ) {
    case $evType.ADD_TEST: return this.addTest( ev.test );
    case $evType.START_RUN: return this.startRun();
    case $evType.START_TEST: return this.startTest( ev.test );
    case $evType.COMPLETE_TEST: return this.completeTest( ev.test, ev.err );
    case $evType.COMPLETE_RUN: return this.completeRun();
    }
}

$mod.addHtmlReporter = function( opts ) {
    
    var tr = $inputs.hasKey( opts, "testRun", "opts" );

    var rpt = new HtmlReporter( opts );
    rpt.init( tr );
};

});
