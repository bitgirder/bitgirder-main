package reactor

type DepthTracker struct {
    depth int
}

func NewDepthTracker() *DepthTracker { return &DepthTracker{} }

func ( t *DepthTracker ) Depth() int { return t.depth }

func ( t *DepthTracker ) ProcessEvent( ev ReactorEvent ) error {
    switch ev.( type ) {
    case *MapStartEvent, *StructStartEvent, *ListStartEvent: t.depth++
    case *EndEvent: t.depth--
    }
    return nil
}
