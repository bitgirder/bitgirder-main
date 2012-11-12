require 'bitgirder/core'
require 'mingle'
require 'bitgirder/ops/build'

module BitGirder
module Ops
module Build
module Distro

include BitGirder::Ops::Build

class AbstractDistroTask < TaskExecutor

    attr :dist_def
    
    # overridable
    public
    def dist
        target.path[ 3 ] or raise "No dist set"
    end

    public
    def dist_def_file
        "#{@workspace.root}/#{dist}/distro.json"
    end

    public
    def init

        @dist_def = 
            ws_ctx.load_mingle_struct_file( file_exists( dist_def_file ) )
    end

    # Could be overridden if needed
    public
    def code_type
        target().path[ 0 ]
    end

    public
    def ws_ctx( opts = {} )
        
        not_nil( opts, :opts )
        super( { proj: dist, code_type: code_type }.merge( opts ) )
    end

    public
    def direct_deps( opts = {} )
        
        not_nil( opts, :opts )
        code_type = opts[ :code_type ] || code_type()

        res =
            if ct = dist_def[ code_type ]
                ct[ :direct_deps ] || []
            else
                []
            end
        
        unless opts[ :no_validate_deps ]
            ws_ctx.validate_deps( 
                deps: res, 
                code_type: code_type,
                err_suffix: "(#{code_type} deps in dist '#{dist}')" 
            )
        end

        res
    end

    public
    def dist_build_dir
        ws_ctx.proj_build_dir
    end
end

end
end
end
end
