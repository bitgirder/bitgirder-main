package service

import (
    mgRct "mingle/reactor"
)

type proxyProc struct {
    proc mgRct.ReactorEventProcessor
    depth int
}

func ( p *proxyProc ) isDone() bool { return p.depth == 0 }

func ( p *proxyProc ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    switch ev.( type ) {
    case *mgRct.ListStartEvent, *mgRct.MapStartEvent, *mgRct.StructStartEvent:
        p.depth++
    case *mgRct.EndEvent: p.depth--
    }
    return p.proc.ProcessEvent( ev )
}
