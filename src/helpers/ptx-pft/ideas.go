/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// pft ideas — Opportunity Management v2 (Venture model) authoring surface.
// Authors and derives a self-contained '.venture' workspace: the discovery
// spine (Venture -> Initiative -> Idea) and the delivery spine (Team/Project ->
// Epic -> Use-Case), validating every write against the api #015 contract.
// This is functionally separate from the feedback-tool sync namespace.

// handleIdeasCommand dispatches the 'pft ideas' subcommands.
func handleIdeasCommand(args []string) {
	if len(args) == 0 {
		showIdeasHelp()
		return
	}
	sub := args[0]
	rest := args[1:]

	switch sub {
	case "new-venture":
		handleNewVenture(rest)
	case "add-product":
		handleAddProduct(rest)
	case "add-team":
		handleAddTeam(rest)
	case "add-initiative":
		handleAddInitiative(rest)
	case "initiative-link":
		handleInitiativeLink(rest)
	case "add-idea", "add":
		handleAddIdea(rest)
	case "complexity":
		handleComplexity(rest)
	case "add-usecase":
		handleAddUseCase(rest)
	case "new-backlog":
		handleNewBacklog(rest)
	case "add-epic":
		handleAddEpic(rest)
	case "epic-link":
		handleEpicLink(rest)
	case "estimate":
		handleEstimate(rest)
	case "link":
		handleIdeasLink(rest)
	case "graph":
		handleIdeasGraph(rest)
	case "rollup":
		handleRollup(rest)
	case "list":
		handleIdeasList(rest)
	case "show":
		handleIdeasShow(rest)
	case "migrate":
		handleIdeasMigrate(rest)
	case "--help", "-h":
		showIdeasHelp()
	default:
		fmt.Printf("Unknown pft ideas subcommand: %s\n", sub)
		fmt.Println("Run 'portunix pft ideas --help' for available commands")
	}
}

// parseIdeaArgs splits args into positional values and a --flag map. A flag
// with no following value (or followed by another --flag) is recorded as a
// boolean with an empty value.
func parseIdeaArgs(args []string) ([]string, map[string]string) {
	var positionals []string
	flags := map[string]string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--") {
			key := a[2:]
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				flags[key] = args[i+1]
				i++
			} else {
				flags[key] = ""
			}
		} else {
			positionals = append(positionals, a)
		}
	}
	return positionals, flags
}

// ideaFail prints an error and marks it as a command failure line.
func ideaFail(format string, a ...interface{}) {
	fmt.Printf("✗ "+format+"\n", a...)
}

// firstArg returns the first positional argument or "".
func firstArg(pos []string) string {
	if len(pos) > 0 {
		return pos[0]
	}
	return ""
}

// entityExists reports whether the file at path is present.
func entityExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ---------------------------------------------------------------------------
// Container / registry
// ---------------------------------------------------------------------------

func handleNewVenture(args []string) {
	pos, flags := parseIdeaArgs(args)
	name := firstArg(pos)
	if err := validateSlug(name); err != nil {
		ideaFail("new-venture: %v", err)
		fmt.Println("Usage: portunix pft ideas new-venture <name> [--name \"...\"] [--path <parent-dir>]")
		return
	}

	parent, _ := filepath.Abs(".")
	if flags["path"] != "" {
		parent, _ = filepath.Abs(flags["path"])
	}
	root := filepath.Join(parent, name+".venture")
	if entityExists(venturePath(root)) {
		ideaFail("venture already exists: %s", root)
		return
	}

	// Skeleton directories.
	for _, d := range []string{productsDir(root), teamsDir(root), initiativesDir(root), recordsDir(root), backlogsDir(root)} {
		if err := os.MkdirAll(d, 0755); err != nil {
			ideaFail("cannot create %s: %v", d, err)
			return
		}
	}

	display := flags["name"]
	if display == "" {
		display = name
	}
	ts := nowTS()
	ven := &Venture{
		SchemaVersion: contentSchemaVersion,
		ID:            name,
		Slug:          name,
		Name:          display,
		Description:   flags["description"],
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}
	if err := writeEntity(venturePath(root), ven); err != nil {
		ideaFail("cannot write venture.json: %v", err)
		return
	}

	// Default (empty) discovery backlog so ideas have a home to register in.
	bk := &Backlog{
		ID:        name,
		Slug:      name,
		Name:      display + " discovery backlog",
		Kind:      "discovery",
		CreatedAt: ts,
		UpdatedAt: ts,
	}
	if err := writeEntity(backlogPath(root, name), bk); err != nil {
		ideaFail("cannot write backlog: %v", err)
		return
	}

	fmt.Printf("✓ Created venture %s\n", root)
	fmt.Printf("  venture.json + backlogs/%s.backlog.json (empty discovery backlog)\n", name)
}

func handleAddProduct(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if err := validateSlug(slug); err != nil {
		ideaFail("add-product: %v", err)
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	name := flags["name"]
	if name == "" {
		name = slug
	}
	ts := nowTS()
	p := &Product{
		ID:        slug,
		Slug:      slug,
		Name:      name,
		Repo:      flags["repo"],
		Component: flags["component"],
		CreatedAt: ts,
		UpdatedAt: ts,
	}
	if err := writeEntity(productPath(root, slug), p); err != nil {
		ideaFail("add-product: %v", err)
		return
	}
	fmt.Printf("✓ Added product %s (products/%s.product.json)\n", slug, slug)
}

func handleAddTeam(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if err := validateSlug(slug); err != nil {
		ideaFail("add-team: %v", err)
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	name := flags["name"]
	if name == "" {
		name = slug
	}
	kind := flags["kind"]
	if kind == "" {
		kind = "team"
	}
	ts := nowTS()
	t := &Team{
		ID:        slug,
		Slug:      slug,
		Name:      name,
		Kind:      kind,
		CreatedAt: ts,
		UpdatedAt: ts,
	}
	if err := writeEntity(teamPath(root, slug), t); err != nil {
		ideaFail("add-team: %v", err)
		return
	}
	fmt.Printf("✓ Added %s %s (teams/%s.team.json)\n", kind, slug, slug)
}

// ---------------------------------------------------------------------------
// Discovery spine
// ---------------------------------------------------------------------------

func handleAddInitiative(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if err := validateSlug(slug); err != nil {
		ideaFail("add-initiative: %v", err)
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	title := flags["title"]
	if title == "" {
		title = slug
	}
	ts := nowTS()
	ini := &Initiative{
		ID:        slug,
		Slug:      slug,
		Title:     title,
		Approach:  flags["approach"],
		Status:    "active",
		Owner:     flags["owner"],
		CreatedAt: ts,
		UpdatedAt: ts,
	}
	if err := writeEntity(initiativePath(root, slug), ini); err != nil {
		ideaFail("add-initiative: %v", err)
		return
	}
	fmt.Printf("✓ Added initiative %s (initiatives/%s.initiative.json)\n", slug, slug)
}

func handleAddIdea(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if err := validateSlug(slug); err != nil {
		ideaFail("add-idea: %v", err)
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	title := flags["title"]
	if title == "" {
		title = slug
	}
	ts := nowTS()
	o := &Opportunity{
		SchemaVersion: contentSchemaVersion,
		ID:            slug,
		Slug:          slug,
		Title:         title,
		Origin:        flags["origin"],
		Status:        "new",
		Domain:        flags["domain"],
		Author:        flags["author"],
		Complexity:    nil, // coarse complexity, assessed later; never story points
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}
	if err := writeEntity(opportunityPath(root, slug), o); err != nil {
		ideaFail("add-idea: %v", err)
		return
	}
	fmt.Printf("✓ Added idea %s (records/%s.opportunity.json)\n", slug, slug)

	// Register the idea as a node in the discovery backlog.
	if bk, err := findDiscoveryBacklog(root); err == nil {
		bk.MemberRefs = appendUnique(bk.MemberRefs, slug)
		bk.UpdatedAt = ts
		if err := writeEntity(backlogPath(root, bk.Slug), bk); err != nil {
			ideaFail("could not update discovery backlog: %v", err)
			return
		}
		fmt.Printf("  registered in discovery backlog %q\n", bk.Slug)
	} else {
		fmt.Printf("  (not registered in a backlog: %v)\n", err)
	}
}

func handleComplexity(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if slug == "" {
		ideaFail("complexity: idea slug required")
		fmt.Println("Usage: portunix pft ideas complexity <idea> --set xs|s|m|l|xl")
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	set := flags["set"]
	if err := requireEnum("complexity", set, enumComplexity); err != nil {
		ideaFail("%v", err)
		return
	}
	path := opportunityPath(root, slug)
	o := &Opportunity{}
	if err := readEntity(path, o); err != nil {
		ideaFail("complexity: idea %q not found (%v)", slug, err)
		return
	}
	o.Complexity = strp(set)
	o.UpdatedAt = nowTS()
	if err := writeEntity(path, o); err != nil {
		ideaFail("complexity: %v", err)
		return
	}
	fmt.Printf("✓ Set complexity of %s to %s\n", slug, set)
}

func handleInitiativeLink(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if slug == "" {
		ideaFail("initiative-link: initiative slug required")
		fmt.Println("Usage: portunix pft ideas initiative-link <initiative> --idea <idea>")
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	idea := flags["idea"]
	if idea == "" {
		ideaFail("initiative-link: --idea <idea> required")
		return
	}
	if !entityExists(opportunityPath(root, idea)) {
		ideaFail("initiative-link: idea %q not found", idea)
		return
	}
	path := initiativePath(root, slug)
	ini := &Initiative{}
	if err := readEntity(path, ini); err != nil {
		ideaFail("initiative-link: initiative %q not found (%v)", slug, err)
		return
	}
	ini.IdeaRefs = appendUnique(ini.IdeaRefs, idea)
	ini.UpdatedAt = nowTS()
	if err := writeEntity(path, ini); err != nil {
		ideaFail("initiative-link: %v", err)
		return
	}
	fmt.Printf("✓ Linked idea %s to initiative %s\n", idea, slug)
}

// ---------------------------------------------------------------------------
// Delivery spine
// ---------------------------------------------------------------------------

func handleAddUseCase(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if err := validateSlug(slug); err != nil {
		ideaFail("add-usecase: %v", err)
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	product := flags["product"]
	if product == "" {
		ideaFail("add-usecase: --product <product> required")
		return
	}
	if !entityExists(productPath(root, product)) {
		ideaFail("add-usecase: product %q not found", product)
		return
	}
	var implements []string
	for _, idea := range splitCSV(flags["implements"]) {
		if !entityExists(opportunityPath(root, idea)) {
			ideaFail("add-usecase: idea %q not found", idea)
			return
		}
		implements = appendUnique(implements, idea)
	}
	title := flags["title"]
	if title == "" {
		title = slug
	}
	ts := nowTS()
	u := &UseCase{
		ID:                 slug,
		Slug:               slug,
		Title:              title,
		Actor:              flags["actor"],
		Goal:               flags["goal"],
		Status:             "new",
		StoryPoints:        nil, // real story points arrive via 'estimate'
		ProductRef:         product,
		ImplementsIdeaRefs: implements,
	}
	u.CreatedAt = ts
	u.UpdatedAt = ts
	if err := writeEntity(usecasePath(root, slug), u); err != nil {
		ideaFail("add-usecase: %v", err)
		return
	}
	fmt.Printf("✓ Added use-case %s (records/%s.usecase.json)\n", slug, slug)
	fmt.Printf("  productRef=%s implements=%v\n", product, implements)
}

func handleNewBacklog(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if err := validateSlug(slug); err != nil {
		ideaFail("new-backlog: %v", err)
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	kind := flags["kind"]
	if err := requireEnum("backlog kind", kind, enumBacklogKind); err != nil {
		ideaFail("%v", err)
		fmt.Println("Usage: portunix pft ideas new-backlog <slug> --kind discovery|delivery [--team <team>]")
		return
	}
	var teamRef *string
	if team := flags["team"]; team != "" {
		if !entityExists(teamPath(root, team)) {
			ideaFail("new-backlog: team %q not found", team)
			return
		}
		teamRef = strp(team)
	}
	name := flags["name"]
	if name == "" {
		name = slug
	}
	ts := nowTS()
	b := &Backlog{
		ID:        slug,
		Slug:      slug,
		Name:      name,
		Kind:      kind,
		TeamRef:   teamRef,
		CreatedAt: ts,
		UpdatedAt: ts,
	}
	if err := writeEntity(backlogPath(root, slug), b); err != nil {
		ideaFail("new-backlog: %v", err)
		return
	}
	fmt.Printf("✓ Created %s backlog %s (backlogs/%s.backlog.json)\n", kind, slug, slug)
}

func handleAddEpic(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if err := validateSlug(slug); err != nil {
		ideaFail("add-epic: %v", err)
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	backlog := flags["backlog"]
	if backlog == "" {
		ideaFail("add-epic: --backlog <backlog> required")
		return
	}
	bkPath := backlogPath(root, backlog)
	bk := &Backlog{}
	if err := readEntity(bkPath, bk); err != nil {
		ideaFail("add-epic: backlog %q not found (%v)", backlog, err)
		return
	}
	title := flags["title"]
	if title == "" {
		title = slug
	}
	var initRefs []string
	if ini := flags["initiative"]; ini != "" {
		if !entityExists(initiativePath(root, ini)) {
			ideaFail("add-epic: initiative %q not found", ini)
			return
		}
		initRefs = appendUnique(initRefs, ini)
	}
	ts := nowTS()
	e := &Epic{
		ID:             slug,
		Slug:           slug,
		Title:          title,
		Status:         "new",
		BacklogRef:     backlog,
		InitiativeRefs: initRefs,
		CreatedAt:      ts,
		UpdatedAt:      ts,
	}
	if err := writeEntity(epicPath(root, slug), e); err != nil {
		ideaFail("add-epic: %v", err)
		return
	}
	bk.MemberRefs = appendUnique(bk.MemberRefs, slug)
	bk.UpdatedAt = ts
	if err := writeEntity(bkPath, bk); err != nil {
		ideaFail("add-epic: could not update backlog: %v", err)
		return
	}
	fmt.Printf("✓ Added epic %s to backlog %s (backlogs/%s.epic.json)\n", slug, backlog, slug)
}

func handleEpicLink(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if slug == "" {
		ideaFail("epic-link: epic slug required")
		fmt.Println("Usage: portunix pft ideas epic-link <epic> --usecase <usecase>")
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	uc := flags["usecase"]
	if uc == "" {
		ideaFail("epic-link: --usecase <usecase> required")
		return
	}
	if !entityExists(usecasePath(root, uc)) {
		ideaFail("epic-link: use-case %q not found", uc)
		return
	}
	ePath := epicPath(root, slug)
	e := &Epic{}
	if err := readEntity(ePath, e); err != nil {
		ideaFail("epic-link: epic %q not found (%v)", slug, err)
		return
	}
	ts := nowTS()
	e.UseCaseRefs = appendUnique(e.UseCaseRefs, uc)
	e.UpdatedAt = ts
	if err := writeEntity(ePath, e); err != nil {
		ideaFail("epic-link: %v", err)
		return
	}
	// Keep the use-case visible in the epic's backlog membership too.
	if e.BacklogRef != "" {
		bkPath := backlogPath(root, e.BacklogRef)
		bk := &Backlog{}
		if err := readEntity(bkPath, bk); err == nil {
			bk.MemberRefs = appendUnique(bk.MemberRefs, uc)
			bk.UpdatedAt = ts
			_ = writeEntity(bkPath, bk)
		}
	}
	fmt.Printf("✓ Linked use-case %s to epic %s\n", uc, slug)
}

func handleEstimate(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if slug == "" {
		ideaFail("estimate: use-case slug required")
		fmt.Println("Usage: portunix pft ideas estimate <usecase> --by <who> --sp <n> --why \"...\"")
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	by := flags["by"]
	if by == "" {
		ideaFail("estimate: --by <who> required")
		return
	}
	if flags["sp"] == "" {
		ideaFail("estimate: --sp <n> required")
		return
	}
	sp, err := strconv.ParseFloat(flags["sp"], 64)
	if err != nil {
		ideaFail("estimate: --sp must be a number (got %q)", flags["sp"])
		return
	}

	ucPath := usecasePath(root, slug)
	u := &UseCase{}
	if err := readEntity(ucPath, u); err != nil {
		ideaFail("estimate: use-case %q not found (%v)", slug, err)
		return
	}

	// Append the round to the estimation history (create it if missing).
	estPath := estimationPath(root, slug)
	est := &Estimation{SchemaVersion: contentSchemaVersion, SubjectRef: u.ID}
	if entityExists(estPath) {
		if err := readEntity(estPath, est); err != nil {
			ideaFail("estimate: cannot read %s (%v)", estPath, err)
			return
		}
	}
	ts := nowTS()
	est.Estimations = append(est.Estimations, EstimationRound{
		Estimator:     by,
		Argumentation: flags["why"],
		StoryPoints:   sp,
		Timestamp:     ts,
	})
	if err := writeEntity(estPath, est); err != nil {
		ideaFail("estimate: %v", err)
		return
	}

	// Roll the latest estimate up onto the use-case.
	u.StoryPoints = f64p(sp)
	u.UpdatedAt = ts
	if err := writeEntity(ucPath, u); err != nil {
		ideaFail("estimate: %v", err)
		return
	}
	fmt.Printf("✓ Estimated %s at %s SP by %s (rolled up to use-case.storyPoints)\n", slug, trimFloat(sp), by)
}

func handleIdeasLink(args []string) {
	pos, flags := parseIdeaArgs(args)
	if len(pos) < 2 {
		ideaFail("link: two entity slugs required")
		fmt.Println("Usage: portunix pft ideas link <a> <b> --type inspired-by|duplicate|split-into|merged-with|implements [--backlog <slug>]")
		return
	}
	a, b := pos[0], pos[1]
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	rel := flags["type"]
	if err := requireEnum("link type", rel, enumRelation); err != nil {
		ideaFail("%v", err)
		return
	}

	// Locate the backlog that owns the relation.
	var bk *Backlog
	if name := flags["backlog"]; name != "" {
		bk = &Backlog{}
		if err := readEntity(backlogPath(root, name), bk); err != nil {
			ideaFail("link: backlog %q not found (%v)", name, err)
			return
		}
	} else {
		kind := "discovery"
		if rel == "implements" {
			kind = "delivery"
		}
		bk, err = findBacklogOfKind(root, kind)
		if err != nil {
			ideaFail("link: %v (pass --backlog <slug>)", err)
			return
		}
	}

	ts := nowTS()
	bk.Relations = append(bk.Relations, Relation{Source: a, Target: b, Rel: rel})
	bk.UpdatedAt = ts
	if err := writeEntity(backlogPath(root, bk.Slug), bk); err != nil {
		ideaFail("link: %v", err)
		return
	}

	// For 'implements', keep the entity-level truth on the use-case in sync so
	// derived graphs remain correct regardless of the backlog scanned.
	if rel == "implements" {
		ucPath := usecasePath(root, a)
		u := &UseCase{}
		if err := readEntity(ucPath, u); err == nil {
			u.ImplementsIdeaRefs = appendUnique(u.ImplementsIdeaRefs, b)
			u.UpdatedAt = ts
			_ = writeEntity(ucPath, u)
		}
	}
	fmt.Printf("✓ Linked %s -[%s]-> %s in backlog %q\n", a, rel, b, bk.Slug)
}

// findBacklogOfKind returns the sole backlog of the given kind.
func findBacklogOfKind(root, kind string) (*Backlog, error) {
	backlogs, err := loadBacklogs(root)
	if err != nil {
		return nil, err
	}
	var found []*Backlog
	for _, b := range backlogs {
		if b.Kind == kind {
			found = append(found, b)
		}
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("no %s backlog in venture", kind)
	case 1:
		return found[0], nil
	default:
		return nil, fmt.Errorf("multiple %s backlogs", kind)
	}
}

// ---------------------------------------------------------------------------
// Relations / derived
// ---------------------------------------------------------------------------

func handleIdeasGraph(args []string) {
	_, flags := parseIdeaArgs(args)
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	idx, err := loadVentureIndex(root)
	if err != nil {
		ideaFail("graph: %v", err)
		return
	}

	var doc *GraphDoc
	var defaultStem string
	var backlog *Backlog

	switch {
	case flags["backlog"] != "":
		bk := &Backlog{}
		if err := readEntity(backlogPath(root, flags["backlog"]), bk); err != nil {
			ideaFail("graph: backlog %q not found (%v)", flags["backlog"], err)
			return
		}
		backlog = bk
		doc = buildBacklogGraph(idx, bk)
		defaultStem = bk.Slug
	case flags["scope"] == "discovery":
		doc = buildDiscoveryGraph(idx)
		defaultStem = "discovery"
	default:
		ideaFail("graph: specify --backlog <slug> or --scope discovery")
		return
	}

	outPath := flags["out"]
	if outPath == "" {
		outPath = glensPath(root, defaultStem)
	}
	if err := WriteGraphJSON(doc, outPath); err != nil {
		ideaFail("graph: %v", err)
		return
	}
	fmt.Printf("✓ Graph written to %s\n", outPath)
	fmt.Printf("  nodes: %d, links: %d\n", len(doc.Nodes), len(doc.Links))

	// Persist graphRef back onto the backlog when we projected one.
	if backlog != nil {
		if rel, err := filepath.Rel(backlogsDir(root), outPath); err == nil {
			backlog.GraphRef = strp(rel)
		} else {
			backlog.GraphRef = strp(filepath.Base(outPath))
		}
		backlog.UpdatedAt = nowTS()
		_ = writeEntity(backlogPath(root, backlog.Slug), backlog)
	}

	if _, view := flags["view"]; !view {
		return
	}
	if !graphlensAvailable() {
		fmt.Printf("graphlens plugin not installed — run 'portunix plugin install graphlens', then 'portunix graphlens view %s'\n", outPath)
		return
	}
	if err := runGraphlensView(outPath, 0, false); err != nil {
		ideaFail("graph: launching graphlens view: %v", err)
	}
}

func handleRollup(args []string) {
	_, flags := parseIdeaArgs(args)
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	idx, err := loadVentureIndex(root)
	if err != nil {
		ideaFail("rollup: %v", err)
		return
	}

	switch {
	case flags["epic"] != "":
		e := idx.epics[flags["epic"]]
		if e == nil {
			ideaFail("rollup: epic %q not found", flags["epic"])
			return
		}
		sp := epicStoryPoints(idx, e)
		e.StoryPoints = f64p(sp)
		e.UpdatedAt = nowTS()
		if err := writeEntity(epicPath(root, e.Slug), e); err != nil {
			ideaFail("rollup: %v", err)
			return
		}
		fmt.Printf("✓ Epic %s: Σ storyPoints = %s (over %d use-cases)\n", e.Slug, trimFloat(sp), len(e.UseCaseRefs))

	case flags["backlog"] != "":
		bk := idx.backlogs[flags["backlog"]]
		if bk == nil {
			ideaFail("rollup: backlog %q not found", flags["backlog"])
			return
		}
		var total float64
		for _, mid := range bk.MemberRefs {
			if e := idx.epics[mid]; e != nil {
				sp := epicStoryPoints(idx, e)
				e.StoryPoints = f64p(sp)
				e.UpdatedAt = nowTS()
				_ = writeEntity(epicPath(root, e.Slug), e)
				total += sp
			}
		}
		fmt.Printf("✓ Backlog %s: Σ storyPoints = %s\n", bk.Slug, trimFloat(total))

	case flags["initiative"] != "":
		ini := idx.inits[flags["initiative"]]
		if ini == nil {
			ideaFail("rollup: initiative %q not found", flags["initiative"])
			return
		}
		sp := initiativeStoryPoints(idx, ini.ID)
		counts, weighted := complexityBreakdown(idx, ini)
		fmt.Printf("✓ Initiative %s\n", ini.Slug)
		fmt.Printf("  Σ storyPoints (delivery): %s\n", trimFloat(sp))
		fmt.Printf("  Σ complexity  (discovery): %s  [weighted %s]\n", formatComplexityBreakdown(counts), trimFloat(weighted))

	default:
		ideaFail("rollup: specify --epic <e>, --backlog <b>, or --initiative <i>")
	}
}

// ---------------------------------------------------------------------------
// Listing / show
// ---------------------------------------------------------------------------

func handleIdeasList(args []string) {
	_, flags := parseIdeaArgs(args)
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	kind := flags["kind"]

	printIdeas := func() {
		ideas, _ := loadOpportunities(root)
		if len(ideas) == 0 {
			return
		}
		fmt.Println("Ideas:")
		for _, o := range ideas {
			cx := "-"
			if o.Complexity != nil {
				cx = *o.Complexity
			}
			fmt.Printf("  %-24s %-10s complexity=%s  %s\n", o.Slug, o.Status, cx, o.Title)
		}
	}
	printUseCases := func() {
		ucs, _ := loadUseCases(root)
		if len(ucs) == 0 {
			return
		}
		fmt.Println("Use-cases:")
		for _, u := range ucs {
			sp := "-"
			if u.StoryPoints != nil {
				sp = trimFloat(*u.StoryPoints)
			}
			fmt.Printf("  %-24s %-12s SP=%-4s product=%s  %s\n", u.Slug, u.Status, sp, u.ProductRef, u.Title)
		}
	}
	printEpics := func() {
		epics, _ := loadEpics(root)
		if len(epics) == 0 {
			return
		}
		fmt.Println("Epics:")
		for _, e := range epics {
			fmt.Printf("  %-24s backlog=%-16s ucs=%d  %s\n", e.Slug, e.BacklogRef, len(e.UseCaseRefs), e.Title)
		}
	}
	printInitiatives := func() {
		inits, _ := loadInitiatives(root)
		if len(inits) == 0 {
			return
		}
		fmt.Println("Initiatives:")
		for _, i := range inits {
			fmt.Printf("  %-24s %-10s ideas=%d  %s\n", i.Slug, i.Status, len(i.IdeaRefs), i.Title)
		}
	}

	switch kind {
	case "idea":
		printIdeas()
	case "usecase":
		printUseCases()
	case "epic":
		printEpics()
	case "initiative":
		printInitiatives()
	case "":
		printInitiatives()
		printIdeas()
		printUseCases()
		printEpics()
	default:
		ideaFail("list: unknown --kind %q (idea|usecase|epic|initiative)", kind)
	}
}

func handleIdeasShow(args []string) {
	pos, flags := parseIdeaArgs(args)
	slug := firstArg(pos)
	if slug == "" {
		ideaFail("show: entity slug required")
		return
	}
	root, err := resolveVenture(flags["path"])
	if err != nil {
		ideaFail("%v", err)
		return
	}
	// Probe each collection for a file with this slug and print it verbatim.
	candidates := []string{
		opportunityPath(root, slug),
		usecasePath(root, slug),
		epicPath(root, slug),
		initiativePath(root, slug),
		backlogPath(root, slug),
		productPath(root, slug),
		teamPath(root, slug),
		estimationPath(root, slug),
	}
	for _, p := range candidates {
		if entityExists(p) {
			b, err := os.ReadFile(p)
			if err != nil {
				ideaFail("show: %v", err)
				return
			}
			rel, _ := filepath.Rel(root, p)
			fmt.Printf("# %s\n%s", rel, string(b))
			if !strings.HasSuffix(string(b), "\n") {
				fmt.Println()
			}
			return
		}
	}
	ideaFail("show: no entity %q found in venture", slug)
}

// trimFloat renders a story-point number without a trailing ".0".
func trimFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ---------------------------------------------------------------------------
// Migration: .discovery (v1, #196) -> .venture (v2)
// ---------------------------------------------------------------------------

func handleIdeasMigrate(args []string) {
	pos, flags := parseIdeaArgs(args)
	src := firstArg(pos)
	if src == "" {
		ideaFail("migrate: path to a '.discovery' directory required")
		fmt.Println("Usage: portunix pft ideas migrate <name>.discovery")
		return
	}
	srcAbs, _ := filepath.Abs(src)
	if !strings.HasSuffix(srcAbs, ".discovery") {
		ideaFail("migrate: expected a '.discovery' directory, got %s", srcAbs)
		return
	}
	if fi, err := os.Stat(srcAbs); err != nil || !fi.IsDir() {
		ideaFail("migrate: %s is not a directory", srcAbs)
		return
	}
	dst := strings.TrimSuffix(srcAbs, ".discovery") + ".venture"
	if flags["out"] != "" {
		dst, _ = filepath.Abs(flags["out"])
	}
	if entityExists(dst) {
		ideaFail("migrate: target already exists: %s", dst)
		return
	}
	if err := os.Rename(srcAbs, dst); err != nil {
		ideaFail("migrate: cannot rename dir: %v", err)
		return
	}

	// discovery.json -> venture.json (add v2 identity fields).
	slug := strings.TrimSuffix(filepath.Base(dst), ".venture")
	oldMeta := filepath.Join(dst, "discovery.json")
	ts := nowTS()
	ven := &Venture{SchemaVersion: contentSchemaVersion, ID: slug, Slug: slug, Name: slug, CreatedAt: ts, UpdatedAt: ts}
	if entityExists(oldMeta) {
		var raw map[string]interface{}
		if err := readEntity(oldMeta, &raw); err == nil {
			if v, ok := raw["name"].(string); ok && v != "" {
				ven.Name = v
			}
			if v, ok := raw["description"].(string); ok {
				ven.Description = v
			}
		}
		_ = os.Remove(oldMeta)
	}
	if err := writeEntity(venturePath(dst), ven); err != nil {
		ideaFail("migrate: cannot write venture.json: %v", err)
		return
	}

	// Move flat *.opportunity.json into records/ and drop the legacy v1
	// 'estimate.storyPoints' (ideas carry complexity in v2, not story points).
	_ = os.MkdirAll(recordsDir(dst), 0755)
	moved := 0
	entries, _ := os.ReadDir(dst)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".opportunity.json") {
			continue
		}
		p := filepath.Join(dst, e.Name())
		var raw map[string]interface{}
		if err := readEntity(p, &raw); err == nil {
			delete(raw, "estimate") // v1 estimate.storyPoints has no place on a v2 idea
			if _, ok := raw["complexity"]; !ok {
				raw["complexity"] = nil
			}
			target := filepath.Join(recordsDir(dst), e.Name())
			if err := writeEntity(target, raw); err == nil {
				_ = os.Remove(p)
				moved++
			}
		}
	}

	fmt.Printf("✓ Migrated %s -> %s\n", srcAbs, dst)
	fmt.Printf("  venture.json written, %d idea(s) moved into records/ (legacy estimate dropped)\n", moved)
	fmt.Println("  Note: story points now belong to product-bound use-cases — create them with 'add-usecase' + 'estimate'.")
	fmt.Println("  Run 'git add -A' to record the rename.")
}

// ---------------------------------------------------------------------------
// Help
// ---------------------------------------------------------------------------

func showIdeasHelp() {
	fmt.Println("Usage: portunix pft ideas <subcommand> [options]")
	fmt.Println()
	fmt.Println("Opportunity Management v2 (Venture model). Authors a self-contained")
	fmt.Println("'.venture' workspace with a discovery spine (Venture -> Initiative -> Idea)")
	fmt.Println("and a delivery spine (Team/Project -> Epic -> Use-Case).")
	fmt.Println()
	fmt.Println("Container / registry:")
	fmt.Println("  new-venture <name> [--name \"...\"] [--path <dir>]   Create <name>.venture/")
	fmt.Println("  add-product <slug> --name \"...\" [--repo <url>]      Register a product")
	fmt.Println("  add-team <slug> --name \"...\" [--kind team|project]  Register a team/project")
	fmt.Println()
	fmt.Println("Discovery spine:")
	fmt.Println("  add-initiative <slug> --title \"...\" [--approach iso16355|ai-funnel|mixed]")
	fmt.Println("  add-idea <slug> --title \"...\" [--origin ai-funnel|iso16355]   (alias: add)")
	fmt.Println("  complexity <idea> --set xs|s|m|l|xl")
	fmt.Println("  initiative-link <initiative> --idea <idea>")
	fmt.Println()
	fmt.Println("Delivery spine:")
	fmt.Println("  add-usecase <slug> --title \"...\" --product <p> --implements <idea>[,<idea>...]")
	fmt.Println("  new-backlog <slug> --kind discovery|delivery [--team <team>]")
	fmt.Println("  add-epic <slug> --backlog <backlog> --title \"...\" [--initiative <initiative>]")
	fmt.Println("  epic-link <epic> --usecase <usecase>")
	fmt.Println("  estimate <usecase> --by <who> --sp <n> --why \"...\"")
	fmt.Println()
	fmt.Println("Relations / derived:")
	fmt.Println("  link <a> <b> --type inspired-by|duplicate|split-into|merged-with|implements")
	fmt.Println("  graph [--backlog <slug>|--scope discovery] [--out <file>.glens.json] [--view]")
	fmt.Println("  rollup [--initiative <i>|--backlog <b>|--epic <e>]")
	fmt.Println("  list [--kind idea|usecase|epic|initiative] [--path <venture>]")
	fmt.Println("  show <slug>")
	fmt.Println()
	fmt.Println("Migration:")
	fmt.Println("  migrate <name>.discovery       Convert a v1 '.discovery' to a v2 '.venture'")
}
