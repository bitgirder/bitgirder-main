require 'bitgirder/core'

module BitGirder
module Core

module ObjectPathTestMethods

    include BitGirderMethods

    def self.impl_fail( msg )

        raise msg
    end

    def assert_equal_with_format( expct, act )

        expct_str = expct.format
        act_str = act.format

        return if expct_str == act_str

        self.impl_fail( "expected path #{expct_str}, got: #{act_str}" )
    end

    module_function :assert_equal_with_format
end

end
end
