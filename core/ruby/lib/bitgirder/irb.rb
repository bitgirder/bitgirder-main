require 'bitgirder/core'

require 'irb'

module BitGirder
module Irb

include BitGirder::Core

class Session < BitGirderClass

    bg_attr :setup,
            required: false,
            description: 
                "Init block to be run in context of irb main object before"
    
    # This is a slimmed down version of the standard IRB run loop setup and
    # execution, with our own setup and control logic spliced in. See the source
    # for irb.rb in the ruby distro for the original details.
    #
    # One note is that we seem to need to set the global :MAIN_CONTEXT value
    # for things to work correctly. This may not be the case and we may find
    # later that there are ways to correctly use the library for multiple IRBs.
    # That being said, we run interactively, so there's not a compelling need
    # right now to have more than one for us.
    public
    def run
 
        if IRB.conf[ :MAIN_CONTEXT ]
            raise "IRB session already running (a MAIN_CONTEXT is set)?"
        end

        IRB.setup( nil )
        irb = IRB::Irb.new
        IRB.conf[ :MAIN_CONTEXT ] = irb.context

        @setup.call( irb.context.workspace.main ) if @setup

        trap( "SIGINT" ) { irb.signal_handle }

        catch( :IRB_EXIT ) { irb.eval_input }
    end
end

end
end
