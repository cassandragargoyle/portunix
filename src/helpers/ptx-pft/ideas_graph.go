/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
)

// GraphView projection for Opportunity Management v2 backlogs. Reuses the
// graphlens contract types (GraphDoc/GraphNode/GraphLink, #191) and emits a
// '*.glens.json' document consumed by the graphlens plugin. Nodes are ideas /
// use-cases / epics; links are the backlog relations plus the derived
// 'implements' (use-case -> idea) and 'contains' (epic -> use-case) edges.

// ventureIndex holds the entities of a venture indexed by id for graph/rollup.
type ventureIndex struct {
	ideas    map[string]*Opportunity
	useCases map[string]*UseCase
	epics    map[string]*Epic
	backlogs map[string]*Backlog
	inits    map[string]*Initiative
}

// loadVentureIndex reads every collection of the venture into id-keyed maps.
func loadVentureIndex(root string) (*ventureIndex, error) {
	idx := &ventureIndex{
		ideas:    map[string]*Opportunity{},
		useCases: map[string]*UseCase{},
		epics:    map[string]*Epic{},
		backlogs: map[string]*Backlog{},
		inits:    map[string]*Initiative{},
	}
	ideas, err := loadOpportunities(root)
	if err != nil {
		return nil, err
	}
	for _, o := range ideas {
		idx.ideas[o.ID] = o
	}
	ucs, err := loadUseCases(root)
	if err != nil {
		return nil, err
	}
	for _, u := range ucs {
		idx.useCases[u.ID] = u
	}
	epics, err := loadEpics(root)
	if err != nil {
		return nil, err
	}
	for _, e := range epics {
		idx.epics[e.ID] = e
	}
	backlogs, err := loadBacklogs(root)
	if err != nil {
		return nil, err
	}
	for _, b := range backlogs {
		idx.backlogs[b.ID] = b
	}
	inits, err := loadInitiatives(root)
	if err != nil {
		return nil, err
	}
	for _, i := range inits {
		idx.inits[i.ID] = i
	}
	return idx, nil
}

// graphBuilder accumulates nodes/links while de-duplicating node ids and edges.
type graphBuilder struct {
	nodes     []GraphNode
	links     []GraphLink
	seen      map[string]bool
	seenLinks map[string]bool
}

func newGraphBuilder() *graphBuilder {
	return &graphBuilder{seen: map[string]bool{}, seenLinks: map[string]bool{}}
}

func (g *graphBuilder) addNode(id, typ, label string) {
	if id == "" || g.seen[id] {
		return
	}
	g.seen[id] = true
	g.nodes = append(g.nodes, GraphNode{ID: id, Type: typ, Label: label})
}

func (g *graphBuilder) addLink(source, target, rel string) {
	key := source + "\x00" + target + "\x00" + rel
	if g.seenLinks[key] {
		return
	}
	g.seenLinks[key] = true
	g.links = append(g.links, GraphLink{Source: source, Target: target, Rel: rel})
}

func (idx *ventureIndex) ideaNode(g *graphBuilder, id string) {
	label := id
	if o := idx.ideas[id]; o != nil {
		label = o.Title
	}
	g.addNode(id, "idea", label)
}

func (idx *ventureIndex) useCaseNode(g *graphBuilder, id string) {
	label := id
	if u := idx.useCases[id]; u != nil {
		label = u.Title
	}
	g.addNode(id, "use-case", label)
}

// buildBacklogGraph projects a single backlog into a GraphView document.
func buildBacklogGraph(idx *ventureIndex, b *Backlog) *GraphDoc {
	g := newGraphBuilder()

	if b.Kind == "delivery" {
		// Epics -> use-cases (contains) -> ideas (implements).
		for _, mid := range b.MemberRefs {
			if e := idx.epics[mid]; e != nil {
				g.addNode(e.ID, "epic", e.Title)
				for _, ucID := range e.UseCaseRefs {
					idx.useCaseNode(g, ucID)
					g.addLink(e.ID, ucID, "contains")
					if u := idx.useCases[ucID]; u != nil {
						for _, ideaID := range u.ImplementsIdeaRefs {
							idx.ideaNode(g, ideaID)
							g.addLink(ucID, ideaID, "implements")
						}
					}
				}
			} else if u := idx.useCases[mid]; u != nil {
				idx.useCaseNode(g, mid)
				for _, ideaID := range u.ImplementsIdeaRefs {
					idx.ideaNode(g, ideaID)
					g.addLink(mid, ideaID, "implements")
				}
			}
		}
	} else {
		// Discovery backlog: member ideas.
		for _, mid := range b.MemberRefs {
			idx.ideaNode(g, mid)
		}
	}

	// Explicit member relations recorded on the backlog.
	for _, r := range b.Relations {
		if r.Rel == "implements" {
			idx.useCaseNode(g, r.Source)
			idx.ideaNode(g, r.Target)
		}
		g.addLink(r.Source, r.Target, r.Rel)
	}

	return &GraphDoc{
		Meta:  map[string]interface{}{"title": b.Name},
		Nodes: g.nodes,
		Links: g.links,
	}
}

// buildDiscoveryGraph projects the discovery scope: every idea plus the
// use-cases that implement them, with the 'implements' edges, so the strategy
// spine is visible with its delivery attachments (issue #198).
func buildDiscoveryGraph(idx *ventureIndex) *GraphDoc {
	g := newGraphBuilder()

	for id, o := range idx.ideas {
		g.addNode(id, "idea", o.Title)
	}
	for id, u := range idx.useCases {
		for _, ideaID := range u.ImplementsIdeaRefs {
			if idx.ideas[ideaID] == nil {
				continue
			}
			idx.useCaseNode(g, id)
			g.addLink(id, ideaID, "implements")
		}
	}
	// Idea-to-idea relations from any discovery backlog.
	for _, b := range idx.backlogs {
		if b.Kind != "discovery" {
			continue
		}
		for _, r := range b.Relations {
			g.addLink(r.Source, r.Target, r.Rel)
		}
	}

	return &GraphDoc{
		Meta:  map[string]interface{}{"title": "Discovery — ideas & implementing use-cases"},
		Nodes: g.nodes,
		Links: g.links,
	}
}

// ---------------------------------------------------------------------------
// Rollups
// ---------------------------------------------------------------------------

// tshirtWeight maps a t-shirt complexity to a coarse numeric weight so that a
// set of ideas can be aggregated into a single indicative figure.
var tshirtWeight = map[string]float64{"xs": 1, "s": 2, "m": 3, "l": 5, "xl": 8}

// epicStoryPoints sums the latest story points of the use-cases an epic groups.
func epicStoryPoints(idx *ventureIndex, e *Epic) float64 {
	var sum float64
	for _, ucID := range e.UseCaseRefs {
		if u := idx.useCases[ucID]; u != nil && u.StoryPoints != nil {
			sum += *u.StoryPoints
		}
	}
	return sum
}

// initiativeStoryPoints sums story points across every epic that serves the
// initiative (delivery-spine rollup).
func initiativeStoryPoints(idx *ventureIndex, initiativeID string) float64 {
	var sum float64
	for _, e := range idx.epics {
		for _, ref := range e.InitiativeRefs {
			if ref == initiativeID {
				sum += epicStoryPoints(idx, e)
				break
			}
		}
	}
	return sum
}

// complexityBreakdown tallies the t-shirt complexities of the ideas an
// initiative pursues, returning the per-size counts and a weighted total
// (discovery-spine rollup).
func complexityBreakdown(idx *ventureIndex, ini *Initiative) (map[string]int, float64) {
	counts := map[string]int{}
	var weighted float64
	for _, ideaID := range ini.IdeaRefs {
		o := idx.ideas[ideaID]
		if o == nil || o.Complexity == nil {
			continue
		}
		size := *o.Complexity
		counts[size]++
		weighted += tshirtWeight[size]
	}
	return counts, weighted
}

// formatComplexityBreakdown renders the per-size tally as "xs:1 m:2".
func formatComplexityBreakdown(counts map[string]int) string {
	out := ""
	for _, size := range enumComplexity {
		if n := counts[size]; n > 0 {
			if out != "" {
				out += " "
			}
			out += fmt.Sprintf("%s:%d", size, n)
		}
	}
	if out == "" {
		return "(none)"
	}
	return out
}
