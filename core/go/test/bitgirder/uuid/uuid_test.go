package uuid

import (
    "testing"
    "regexp"
)

var uuidPat *regexp.Regexp

func init() {
    uuidPat = regexp.MustCompile( 
        "^[[:xdigit:]]{8}-" +
        "[[:xdigit:]]{4}-" +
        "[[:xdigit:]]{4}-" +
        "[[:xdigit:]]{4}-" +
        "[[:xdigit:]]{12}$",
    )
}

func TestType4Basic( t *testing.T ) {
    id := MustType4()
    if ! uuidPat.MatchString( id ) {
        t.Fatalf( "%s does not match uuid pattern %s", id, uuidPat )
    }
}

// Not an exhaustive test for collisions or anything like that, more just a
// safeguard against some regression that might cause the Type4() method to
// generate correctly formatted but static or repeating values (for instance,
// 00000000-0000-0000-0000-000000000000)
func TestType4Sanity( t *testing.T ) {
    id1 := MustType4()
    id2 := MustType4()
    if id1 == id2 { t.Fatalf( "id1 and id2 are both %s", id1 ) }
}
