require 'bitgirder/testing'
require 'bitgirder/mysql'

module BitGirder
module MySql

class QuotingTests

    include TestClassMixin

    def test_quote_no_escape
        { 
            "hello" => "hello",
            "go\nhome" => 'go\nhome',
            "quote'this" => "quote\\'this",
        }.each_pair do |str, expct|
            assert_equal( expct, MySql.quote( str ) )
        end
    end
end

end
end
