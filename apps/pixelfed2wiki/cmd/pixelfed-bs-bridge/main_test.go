package main

import "testing"

func TestBuildFacets(t *testing.T) {
	text := "Read https://example.com/post (and #Go)."
	facets := buildFacets(text)
	if len(facets) != 2 {
		t.Fatalf("got %d facets, want 2", len(facets))
	}
	if got := text[facets[0].Index.ByteStart:facets[0].Index.ByteEnd]; got != "https://example.com/post" {
		t.Fatalf("link range = %q", got)
	}
	if facets[0].Features[0].RichtextFacet_Link == nil || facets[0].Features[0].RichtextFacet_Link.Uri != "https://example.com/post" {
		t.Fatalf("unexpected link feature: %#v", facets[0].Features[0])
	}
	if got := text[facets[1].Index.ByteStart:facets[1].Index.ByteEnd]; got != "#Go" {
		t.Fatalf("tag range = %q", got)
	}
	if facets[1].Features[0].RichtextFacet_Tag == nil || facets[1].Features[0].RichtextFacet_Tag.Tag != "Go" {
		t.Fatalf("unexpected tag feature: %#v", facets[1].Features[0])
	}
}
