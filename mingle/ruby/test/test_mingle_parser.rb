require 'mingle'
require 'mingle/test-support'
require 'bitgirder/core'
require 'bitgirder/io'
require 'bitgirder/testing'

module Mingle

class ParseErrorExpectation < BitGirderClass
    bg_attr :col, :processor => :integer, :validation => :nonnegative
    bg_attr :message
end

class RestrictionErrorExpectation < BitGirderClass
    bg_attr :message
end

PARSE_OVERRIDES = {

    :string => {
        "\"a\\udd1e\\ud834\"" => 'Trailing surrogate with no lead: \uDD1E'
    },

    :identifier => {
        'giving-mixedMessages' => 
            'Illegal start of identifier part: "M" (0x4D)'
    },

    :declared_type_name => { 
        'Bad$Char' => 'Unrecognized token: "$" (0x24)',
        'Bad_Char' => 'Unrecognized token: "_" (0x5F)'
    },

    :qualified_type_name => {
        "ns1@v1/T1#T2" => 'Unrecognized token: "#" (0x23)'
    },

    :type_reference => {
            
        "mingle:core@v1/Float32~[1,2)" => {
            :external_form => 'mingle:core@v1/Float32~[1.0,2.0)'
        },
        "Float32~[1,2)" => {
            :external_form => 'mingle:core@v1/Float32~[1.0,2.0)'
        },
        "mingle:core@v1/Timestamp~[\"2012-01-01T12:00:00Z\",\"2012-01-02T12:00:00Z\"]" => {
            :external_form => 'mingle:core@v1/Timestamp~["2012-01-01T12:00:00.000000000Z","2012-01-02T12:00:00.000000000Z"]'
        },
        "Timestamp~[\"2012-01-01T12:00:00Z\",\"2012-01-02T12:00:00Z\"]" => {
            :external_form => 'mingle:core@v1/Timestamp~["2012-01-01T12:00:00.000000000Z","2012-01-02T12:00:00.000000000Z"]'
        },
        "mingle:core@v1/String ~= \"sdf\"" => 'Unrecognized token: "=" (0x3D)',
        "mingle:core@v1/String~" => "Expected type restriction but found: END",
        "Int32~[1,3}" => {
            :err_message => 'Unexpected char in integer part: "}" (0x7D)',
            :err_col => 11
        },
        "Int32~[-\"abc\",2)" => {
            :err_message => 'Unexpected char in integer part: "\"" (0x22)',
            :err_col => 9
        },
        "Int32~[--3,4)" => {
            :err_message => 'Number has empty or invalid integer part',
            :err_col => 8
        },
        "String~\"ab[a-z\"" => 
            'Invalid regex: premature end of ' +
            ( RUBY_VERSION >= "1.9" || RubyVersions.jruby? ? 
                "char-class" : "regular expression" ) +
            ': /ab[a-z/',
        
        "mingle:core@v1/String~=\"sdf\"" => 'Unrecognized token: "=" (0x3D)',
        "Timestamp~[\"2001-0x-22\",)" => 'invalid date: "2001-0x-22"'
    }
}

class ParseTest < BitGirderClass

    include Testing::AssertMethods

    bg_attr :test_type, :processor => :symbol
    bg_attr :input
    bg_attr :external_form, :required => false
    bg_attr :expect, :required => false
    bg_attr :error, :required => false

    private
    def get_override

        res = ( PARSE_OVERRIDES[ @test_type ] || {} )[ @input ] || {}
        res = { :err_message => res } if res.is_a?( String )

        res
    end

    private
    def expect_token( cls )
        
        lx = MingleLexer.as_instance( @input )
        tok, loc = lx.expect_token
        raise "Expected #{cls} but got #{tok.class}" unless tok.is_a?( cls )
        raise "Trailing input" unless lx.eof?

        tok
    end

    private
    def call_parse
        
        case @test_type
        when :string then expect_token( StringToken )
        when :number then ParsedNumber.parse( @input )
        when :identifier then MingleIdentifier.parse( @input )
        when :namespace then MingleNamespace.parse( @input )
        when :declared_type_name then DeclaredTypeName.parse( @input )
        when :qualified_type_name then QualifiedTypeName.parse( @input )
        when :identified_name then MingleIdentifiedName.parse( @input )
        when :type_reference then MingleTypeReference.parse( @input )
        else raise "Unhandled test type: #@test_type"
        end
    end

    private
    def assert_num_roundtrip( n )
        
        n_str = n.external_form
        lx = MingleLexer.as_instance( n_str )
        
        n2, _ = ParsedNumber.parse( n_str )
        assert_equal( n, n2 )
    end

    private
    def assert_result( res )

        assert_equal( @expect, res )
        assert_equal( @expect.hash, res.hash )
        assert( @expect.eql?( res ) )

        unless @external_form.empty?
            expct = get_override[ :external_form ] || @external_form
            assert_equal( expct, res.external_form ) 
        end

        assert_num_roundtrip( res ) if res.is_a?( ParsedNumber )
    end

    private
    def get_message_expect
 
        msg_expct = get_override[ :err_message ]
        msg_expct ||= @error.message.gsub( /U\+00([[:xdigit:]]{2})/, '0x\1' )
    end

    private
    def assert_parse_error( pe )
        
        assert_equal( get_message_expect, pe.err )

        assert_equal( 1, pe.loc.line )
        assert_equal( get_override[ :err_col ] || @error.col, pe.loc.col )
    end

    private
    def assert_restriction_error( re )
        assert_equal( get_message_expect, re.message )
    end

    private
    def assert_error( e )
        
        case 
        when e.is_a?( MingleParseError ) && 
             @error.is_a?( ParseErrorExpectation )
            assert_parse_error( e )
        when e.is_a?( RestrictionTypeError ) &&
             @error.is_a?( RestrictionErrorExpectation )
            assert_restriction_error( e )
        else raise e
        end
    end

    public
    def call
 
        begin

#            code( "Parsing: #{self.inspect}" )
            res = call_parse

            raise "Got #{res}, but expected error #{@error.inspect}" if @error
            assert_result( res )

        rescue => e
            assert_error( e )
        end
    end
end

class ParseTests < Testing::TestHolder

    QN_PARSE_ERROR_EXPECT = QualifiedTypeName.
        get( :"mingle:parser:testing@v1/ParseErrorExpect" )
    
    QN_RESTRICTION_ERROR_EXPECT = QualifiedTypeName.
        get( :"mingle:parser:testing@v1/RestrictionErrorExpect" )
    
    def as_string_token( s )
        StringToken.new( :val => s.fields.expect_string( :string ) )
    end

    def as_numeric_token( s )
        
        f = s.fields

        ParsedNumber.new(
            :negative => f.expect_boolean( :negative ),
            :num => NumericToken.new(
                :int => f.get_string( :int ),
                :frac => f.get_string( :frac ),
                :exp => f.get_string( :exp ),
                :exp_char => f.get_string( :exp_char )
            )
        )
    end

    def as_identifier( s )
        
        parts = s[ :parts ].map { |p| p.to_s }
        MingleIdentifier.send( :new, :parts => parts )
    end

    def as_namespace( s )
        
        MingleNamespace.send( :new,
            :parts => s[ :parts ].map { |p| as_identifier( p ) },
            :version => as_identifier( s[ :version ] )
        )
    end

    def as_declared_type_name( s )
        
        DeclaredTypeName.send( :new, :name => s.fields.expect_string( :name ) )
    end

    def as_qualified_type_name( s )

        QualifiedTypeName.send( :new,
            :namespace => as_namespace( s[ :namespace ] ),
            :name => as_declared_type_name( s[ :name ] )
        )
    end

    def as_identified_name( s )
        
        MingleIdentifiedName.send( :new,
            :namespace => as_namespace( s[ :namespace ] ),
            :names => s[ :names ].map { |n| as_identifier( n ) }
        )
    end

    def as_atomic_type_reference( s )
        
        AtomicTypeReference.send( :new,
            :name => as_expect_value( s[ :name ] ),
            :restriction => as_expect_value( s[ :restriction ] )
        )
    end

    def as_list_type_reference( s )
        
        ListTypeReference.send( :new,
            :element_type => as_expect_value( s[ :element_type ] ),
            :allows_empty => s.fields.expect_boolean( :allows_empty )
        )
    end

    def as_nullable_type_reference( s )
        
        NullableTypeReference.send( :new,
            :type => as_expect_value( s[ :type ] )
        )
    end

    def as_regex_restriction( s )

        RegexRestriction.
            new( :ext_pattern => s.fields.expect_string( :pattern ) )
    end

    def as_range_restriction( s )

        RangeRestriction.new(
            :max_closed => s.fields.expect_boolean( :max_closed ),
            :max => s[ :max ],
            :min => s[ :min ],
            :min_closed => s.fields.expect_boolean( :min_closed )
        )
    end

    def as_expect_value( s )
        
        return nil unless s

        case s.type.name.to_s
        when "StringToken" then as_string_token( s )
        when "NumericToken" then as_numeric_token( s )
        when "Identifier" then as_identifier( s )
        when "Namespace" then as_namespace( s )
        when "DeclaredTypeName" then as_declared_type_name( s )
        when "QualifiedTypeName" then as_qualified_type_name( s )
        when "IdentifiedName" then as_identified_name( s )
        when "AtomicTypeReference" then as_atomic_type_reference( s )
        when "ListTypeReference" then as_list_type_reference( s )
        when "NullableTypeReference" then as_nullable_type_reference( s )
        when "RegexRestriction" then as_regex_restriction( s )
        when "RangeRestriction" then as_range_restriction( s )
        else raise "unhandled expect type: #{s.type}"
        end
    end

    def as_parse_error_expect( s )
        
        ParseErrorExpectation.new(
            :col => s.fields.expect_int( :col ),
            :message => s.fields.expect_string( :message )
        )
    end

    def as_restriction_error_expect( s )
        
        RestrictionErrorExpectation.new( 
            :message => s.fields.expect_string( :message ) )
    end 

    def as_error( s )
        
        return nil unless s

        case s.type
        when QN_PARSE_ERROR_EXPECT then as_parse_error_expect( s )
        when QN_RESTRICTION_ERROR_EXPECT then as_restriction_error_expect( s )
        else raise "unhandled error: #{s.type}"
        end
    end

    # In addition to testing parser coverage, these tests also give us coverage
    # of external form and equality
    def read_tests
        
        res = {}

        MingleTestStructFile.each_struct_in( "parser-tests.bin" ) do |s|
            
            flds = s.fields

            t = ParseTest.new(
                :test_type => 
                    flds.expect_string( :test_type ).tr( "-", "_" ).to_sym,
                :input => flds.expect_string( :in ),
                :external_form => flds.get_string( :external_form ),
                :expect => as_expect_value( flds.get_mingle_struct( :expect ) ),
                :error => as_error( flds.get_mingle_struct( :err ) )
            )
            
            res[ t.input ] = lambda { |ctx| ctx.complete { t.call } }
        end

        res
    end
    
    invocation_factory :read_tests
end

end
