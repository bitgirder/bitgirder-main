require 'bitgirder/core'
require 'bitgirder/ops/build'

include BitGirder::Core
include BitGirder::Ops::Build
include BitGirder::Io

module BitGirder
module Build
module Ops
module Util

class RubyEnvRunner < StandardTask

    def get_direct_dependencies
        []
    end

    def execute( chain )
        
        rb_ctx = ruby_ctx

        opts = rb_ctx.proc_builder_opts
        
        argv = Array.new( @run_ctx[ :argv_remain ] )
        raise "Need a command" if argv.empty?

        opts[ :cmd ] = argv.shift
        opts[ :argv ] += argv
        opts[ :show_env_in_debug ] = @run_opts.get_boolean( :show_env_in_debug )

        UnixProcessBuilder.new( opts ).exec
    end
end

TaskRegistry.instance.register_path( RubyEnvRunner, :util, :ruby_env, :run )

end
end
end
end
