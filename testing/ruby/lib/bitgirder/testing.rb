require 'bitgirder/core'

module BitGirder
module Testing

ENV_TEST_DATA_PATH = "TEST_DATA_PATH"
ENV_TEST_BIN_PATH = "TEST_BIN_PATH"

def self.find_test_file( name, path )
    
    ( ENV[ path ] || "" ).split( ":" ).
        map { |path| "#{path}/#{name}" }.
        find { |f| File.exists?( f ) } or
        raise "file '#{name}' not found on path #{path}"
end

def self.find_test_data( name )
    self.find_test_file( name, ENV_TEST_DATA_PATH )
end

def self.find_test_command( name )
    self.find_test_file( name, ENV_TEST_BIN_PATH )
end

class AssertionFailure < StandardError; end

module AssertMethods

    include BitGirder::Core::BitGirderMethods

    def fail_test( msg = nil )
        
        msg = msg.call if msg.is_a?( Proc )
        raise AssertionFailure.new( msg )
    end

    def assert( passed, msg = nil )
        fail_test( msg ) unless passed
    end

    def assert_false( val, *argv )
        assert( ! val, *argv )
    end

    def assert_equal( expct, actual, msg = nil )
        
        unless expct == actual
            fail_test( msg || "expct != actual (#{expct} != #{actual})" )
        end
    end

    def assert_equal_meth( expct_obj, actual_obj, meth, msg = nil )
        
        not_nil( meth, :meth )
        assert_equal( expct_obj.send( meth ), actual_obj.send( meth ), msg )
    end

    def assert_equal_i( expct, actual, msg = nil )
        assert_equal_meth( expct, actual, :to_i, msg )
    end

    def assert_equal_s( expct, actual, msg = nil )
        assert_equal_meth( expct, actual, :to_s, msg )
    end

    def assert_nil( val, msg = nil )
        assert( val == nil, msg ||= "Expected nil but got #{val}" )
    end

    def assert_match( pat, str, msg = nil )

        msg ||= lambda { "#{str.inspect} does not match #{pat}" }
        assert( pat.match( str ), msg )
    end

    def get_expect_raised_pat( excpts )
            
        case excpts[ 0 ]
            when String then Regexp.new( /^#{Regexp.quote( excpts.shift )}$/ )
            when Regexp then excpts.shift
            else nil
        end
    end

    def assert_raised( *excpts )
        
        res = nil

        pat = get_expect_raised_pat( excpts )

        begin
            yield
        rescue *excpts => e; # got an expected exception
            res = e
        rescue Exception => e
            raise e
        end

        if pat && res
            pat.match( res.message ) or 
                raise "Message #{res.message.inspect} does not match #{pat}"
        end

        res or fail_test( "Expected raise of one of #{excpts.inspect}" )
    end
end

module TestClassMixin

    require 'set'

    INCLUDED_BY = Set.new

    def self.included( mod )

        INCLUDED_BY << mod
        mod.send( :include, AssertMethods )
    end

    def self.mixed_in_by
        return INCLUDED_BY.to_a
    end

end

end
end
