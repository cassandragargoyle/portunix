/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Opportunity Management v2 data model — Go structs mirroring the api #015
// contract schemas (opportunity/venture, revised to the two-spine "Venture"
// model). Schemas are the source of truth and closed (additionalProperties:
// false), so these structs must not emit fields the schema does not declare.
// Entity identity uses id == slug: the CLI references every entity by its slug,
// which is stored verbatim as the id, keeping *Ref links human-readable.

// ---------------------------------------------------------------------------
// Discovery spine: Venture -> Initiative -> Idea
// ---------------------------------------------------------------------------

// Venture is the top-level, self-contained container (venture.schema.json).
type Venture struct {
	SchemaVersion string  `json:"schemaVersion"`
	TenantID      *string `json:"tenantId"`
	WorkspaceID   *string `json:"workspaceId"`
	ID            string  `json:"id"`
	Slug          string  `json:"slug"`
	Name          string  `json:"name"`
	Description   string  `json:"description,omitempty"`
	CreatedAt     string  `json:"createdAt,omitempty"`
	UpdatedAt     string  `json:"updatedAt,omitempty"`
}

// Product is a stable asset a use-case is bound to (product.schema.json).
type Product struct {
	TenantID    *string `json:"tenantId"`
	WorkspaceID *string `json:"workspaceId"`
	ID          string  `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Repo        string  `json:"repo,omitempty"`
	Component   string  `json:"component,omitempty"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	UpdatedAt   string  `json:"updatedAt,omitempty"`
}

// Team is a team/project registry entry (team.schema.json).
type Team struct {
	TenantID    *string  `json:"tenantId"`
	WorkspaceID *string  `json:"workspaceId"`
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	ProductRefs []string `json:"productRefs,omitempty"`
	CreatedAt   string   `json:"createdAt,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
}

// Period is the optional time box of an initiative.
type Period struct {
	Start *string `json:"start"`
	End   *string `json:"end"`
}

// Initiative is a strategic intent referencing ideas (initiative.schema.json).
type Initiative struct {
	TenantID    *string  `json:"tenantId"`
	WorkspaceID *string  `json:"workspaceId"`
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Approach    string   `json:"approach,omitempty"`
	Status      string   `json:"status,omitempty"`
	Period      *Period  `json:"period,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	IdeaRefs    []string `json:"ideaRefs,omitempty"`
	CreatedAt   string   `json:"createdAt,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
}

// SolutionVariant is a candidate approach recorded on an idea.
type SolutionVariant struct {
	Type     string `json:"type"`
	Desc     string `json:"desc"`
	Selected bool   `json:"selected"`
}

// Scores holds coarse 1-5 evaluation scores (null when not yet scored).
type Scores struct {
	BusinessValue    *int `json:"businessValue"`
	Feasibility      *int `json:"feasibility"`
	DataAvailability *int `json:"dataAvailability"`
	Risk             *int `json:"risk"`
	POCSpeed         *int `json:"pocSpeed"`
}

// Interest is the demand signal ("I want this").
type Interest struct {
	Count  int      `json:"count"`
	Voters []string `json:"voters,omitempty"`
}

// Volunteers is the supply signal ("I would work on this").
type Volunteers struct {
	Count  int      `json:"count"`
	People []string `json:"people,omitempty"`
}

// AIAmbassador is an optional AI persona championing an idea.
type AIAmbassador struct {
	Name       string `json:"name,omitempty"`
	Persona    string `json:"persona,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
	AssignedAt string `json:"assignedAt,omitempty"`
}

// Opportunity is an Idea: a stable, product-agnostic demand asset carrying a
// coarse t-shirt complexity, never story points (opportunity.schema.json v2).
type Opportunity struct {
	SchemaVersion    string            `json:"schemaVersion,omitempty"`
	TenantID         *string           `json:"tenantId"`
	WorkspaceID      *string           `json:"workspaceId"`
	ID               string            `json:"id"`
	Slug             string            `json:"slug"`
	Title            string            `json:"title"`
	Origin           string            `json:"origin,omitempty"`
	Status           string            `json:"status,omitempty"`
	Problem          string            `json:"problem,omitempty"`
	Benefit          string            `json:"benefit,omitempty"`
	AISolution       string            `json:"aiSolution,omitempty"`
	SolutionVariants []SolutionVariant `json:"solutionVariants,omitempty"`
	Scores           *Scores           `json:"scores,omitempty"`
	Complexity       *string           `json:"complexity"`
	Interest         *Interest         `json:"interest,omitempty"`
	Volunteers       *Volunteers       `json:"volunteers,omitempty"`
	AIAmbassador     *AIAmbassador     `json:"aiAmbassador,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Domain           string            `json:"domain,omitempty"`
	Author           string            `json:"author,omitempty"`
	CreatedAt        string            `json:"createdAt,omitempty"`
	UpdatedAt        string            `json:"updatedAt,omitempty"`
}

// ---------------------------------------------------------------------------
// Delivery spine: Team/Project -> Epic -> Use-Case
// ---------------------------------------------------------------------------

// UseCase is a product-bound delivery unit carrying real story points
// (use-case.schema.json).
type UseCase struct {
	TenantID           *string  `json:"tenantId"`
	WorkspaceID        *string  `json:"workspaceId"`
	ID                 string   `json:"id"`
	Slug               string   `json:"slug"`
	Title              string   `json:"title"`
	Actor              string   `json:"actor,omitempty"`
	Goal               string   `json:"goal,omitempty"`
	Status             string   `json:"status,omitempty"`
	StoryPoints        *float64 `json:"storyPoints"`
	ProductRef         string   `json:"productRef"`
	ImplementsIdeaRefs []string `json:"implementsIdeaRefs,omitempty"`
	UCDocRef           *string  `json:"ucDocRef"`
	Tags               []string `json:"tags,omitempty"`
	CreatedAt          string   `json:"createdAt,omitempty"`
	UpdatedAt          string   `json:"updatedAt,omitempty"`
}

// Relation is a member-to-member edge stored in a backlog and mirrored into the
// derived GraphView projection.
type Relation struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Rel    string `json:"rel"`
}

// Backlog is a first-class collection (backlog.schema.json), sibling of teams.
type Backlog struct {
	TenantID    *string    `json:"tenantId"`
	WorkspaceID *string    `json:"workspaceId"`
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Kind        string     `json:"kind"`
	TeamRef     *string    `json:"teamRef"`
	MemberRefs  []string   `json:"memberRefs,omitempty"`
	Relations   []Relation `json:"relations,omitempty"`
	GraphRef    *string    `json:"graphRef"`
	CreatedAt   string     `json:"createdAt,omitempty"`
	UpdatedAt   string     `json:"updatedAt,omitempty"`
}

// Epic groups use-cases inside a backlog (epic.schema.json).
type Epic struct {
	TenantID       *string  `json:"tenantId"`
	WorkspaceID    *string  `json:"workspaceId"`
	ID             string   `json:"id"`
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Status         string   `json:"status,omitempty"`
	BacklogRef     string   `json:"backlogRef"`
	InitiativeRefs []string `json:"initiativeRefs,omitempty"`
	UseCaseRefs    []string `json:"useCaseRefs,omitempty"`
	StoryPoints    *float64 `json:"storyPoints"`
	CreatedAt      string   `json:"createdAt,omitempty"`
	UpdatedAt      string   `json:"updatedAt,omitempty"`
}

// EstimationRound is a single planning-poker round.
type EstimationRound struct {
	Estimator     string  `json:"estimator"`
	Argumentation string  `json:"argumentation,omitempty"`
	StoryPoints   float64 `json:"storyPoints"`
	Timestamp     string  `json:"timestamp"`
}

// Estimation is the planning-poker history for a use-case (estimation.schema.json).
// The latest round sets the use-case storyPoints rollup.
type Estimation struct {
	SchemaVersion string            `json:"schemaVersion,omitempty"`
	SubjectRef    string            `json:"subjectRef"`
	Estimations   []EstimationRound `json:"estimations"`
}

// DiscussionMessage is one ordered thread message.
type DiscussionMessage struct {
	Role      string `json:"role"`
	Text      string `json:"text"`
	Author    string `json:"author,omitempty"`
	Timestamp string `json:"timestamp"`
}

// Discussion is a chat thread attachable to any entity (discussion.schema.json).
type Discussion struct {
	SchemaVersion string              `json:"schemaVersion,omitempty"`
	SubjectRef    string              `json:"subjectRef"`
	SubjectKind   string              `json:"subjectKind"`
	Messages      []DiscussionMessage `json:"messages"`
}

// ---------------------------------------------------------------------------
// Enumerations (mirror the schema enum vocabularies)
// ---------------------------------------------------------------------------

var (
	enumTeamKind        = []string{"team", "project"}
	enumApproach        = []string{"iso16355", "ai-funnel", "mixed"}
	enumInitiativeState = []string{"active", "closed", "archived"}
	enumOrigin          = []string{"ai-funnel", "iso16355"}
	enumIdeaStatus      = []string{"new", "ready", "evaluated", "poc", "done", "rejected"}
	enumComplexity      = []string{"xs", "s", "m", "l", "xl"}
	enumUCStatus        = []string{"new", "ready", "in-progress", "done", "rejected"}
	enumBacklogKind     = []string{"discovery", "delivery"}
	enumRelation        = []string{"inspired-by", "duplicate", "split-into", "merged-with", "implements"}
)

// contentSchemaVersion is stamped into documents that carry a schemaVersion.
const contentSchemaVersion = "2.0.0"

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

var slugRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// validateSlug ensures a slug is present and filesystem-friendly.
func validateSlug(s string) error {
	if s == "" {
		return fmt.Errorf("a slug/name is required")
	}
	if !slugRe.MatchString(s) {
		return fmt.Errorf("invalid slug %q (allowed: letters, digits, '.', '_', '-')", s)
	}
	return nil
}

// enumOK reports whether val is in allowed. An empty val is treated as valid
// (optional field left unset); required-ness is checked separately.
func enumOK(val string, allowed []string) bool {
	if val == "" {
		return true
	}
	for _, a := range allowed {
		if val == a {
			return true
		}
	}
	return false
}

// requireEnum validates a mandatory enum value (empty is rejected).
func requireEnum(field, val string, allowed []string) error {
	if val == "" || !enumOK(val, allowed) {
		return fmt.Errorf("%s must be one of: %s (got %q)", field, strings.Join(allowed, ", "), val)
	}
	return nil
}

// optionalEnum validates an optional enum value (empty is accepted).
func optionalEnum(field, val string, allowed []string) error {
	if !enumOK(val, allowed) {
		return fmt.Errorf("%s must be one of: %s (got %q)", field, strings.Join(allowed, ", "), val)
	}
	return nil
}

// Validate checks required fields and enum vocabularies for each entity, so a
// document that would be rejected by the api #015 JSON Schema is caught before
// it is written to disk.

func (v *Venture) Validate() error {
	if v.SchemaVersion == "" || v.ID == "" || v.Slug == "" || v.Name == "" {
		return fmt.Errorf("venture requires schemaVersion, id, slug, name")
	}
	return nil
}

func (p *Product) Validate() error {
	if p.ID == "" || p.Slug == "" || p.Name == "" {
		return fmt.Errorf("product requires id, slug, name")
	}
	return nil
}

func (t *Team) Validate() error {
	if t.ID == "" || t.Slug == "" || t.Name == "" {
		return fmt.Errorf("team requires id, slug, name")
	}
	return requireEnum("team kind", t.Kind, enumTeamKind)
}

func (i *Initiative) Validate() error {
	if i.ID == "" || i.Slug == "" || i.Title == "" {
		return fmt.Errorf("initiative requires id, slug, title")
	}
	if err := optionalEnum("initiative approach", i.Approach, enumApproach); err != nil {
		return err
	}
	return optionalEnum("initiative status", i.Status, enumInitiativeState)
}

func (o *Opportunity) Validate() error {
	if o.ID == "" || o.Slug == "" || o.Title == "" {
		return fmt.Errorf("idea requires id, slug, title")
	}
	if err := optionalEnum("idea origin", o.Origin, enumOrigin); err != nil {
		return err
	}
	if err := optionalEnum("idea status", o.Status, enumIdeaStatus); err != nil {
		return err
	}
	if o.Complexity != nil {
		if err := requireEnum("idea complexity", *o.Complexity, enumComplexity); err != nil {
			return err
		}
	}
	return nil
}

func (u *UseCase) Validate() error {
	if u.ID == "" || u.Slug == "" || u.Title == "" || u.ProductRef == "" {
		return fmt.Errorf("use-case requires id, slug, title, productRef")
	}
	return optionalEnum("use-case status", u.Status, enumUCStatus)
}

func (b *Backlog) Validate() error {
	if b.ID == "" || b.Slug == "" || b.Name == "" {
		return fmt.Errorf("backlog requires id, slug, name")
	}
	if err := requireEnum("backlog kind", b.Kind, enumBacklogKind); err != nil {
		return err
	}
	for _, r := range b.Relations {
		if err := requireEnum("relation rel", r.Rel, enumRelation); err != nil {
			return err
		}
	}
	return nil
}

func (e *Epic) Validate() error {
	if e.ID == "" || e.Slug == "" || e.Title == "" || e.BacklogRef == "" {
		return fmt.Errorf("epic requires id, slug, title, backlogRef")
	}
	return optionalEnum("epic status", e.Status, enumUCStatus)
}

func (e *Estimation) Validate() error {
	if e.SubjectRef == "" || len(e.Estimations) == 0 {
		return fmt.Errorf("estimation requires subjectRef and at least one round")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Path helpers (venture directory layout, v2.1)
// ---------------------------------------------------------------------------

func venturePath(root string) string    { return filepath.Join(root, "venture.json") }
func productsDir(root string) string    { return filepath.Join(root, "products") }
func teamsDir(root string) string       { return filepath.Join(root, "teams") }
func initiativesDir(root string) string { return filepath.Join(root, "initiatives") }
func recordsDir(root string) string     { return filepath.Join(root, "records") }
func backlogsDir(root string) string    { return filepath.Join(root, "backlogs") }

func productPath(root, slug string) string {
	return filepath.Join(productsDir(root), slug+".product.json")
}
func teamPath(root, slug string) string { return filepath.Join(teamsDir(root), slug+".team.json") }
func initiativePath(root, slug string) string {
	return filepath.Join(initiativesDir(root), slug+".initiative.json")
}
func opportunityPath(root, slug string) string {
	return filepath.Join(recordsDir(root), slug+".opportunity.json")
}
func usecasePath(root, slug string) string {
	return filepath.Join(recordsDir(root), slug+".usecase.json")
}
func estimationPath(root, slug string) string {
	return filepath.Join(recordsDir(root), slug+".estimation.json")
}
func discussionPath(root, slug string) string {
	return filepath.Join(recordsDir(root), slug+".discussion.json")
}
func backlogPath(root, slug string) string {
	return filepath.Join(backlogsDir(root), slug+".backlog.json")
}
func epicPath(root, slug string) string  { return filepath.Join(backlogsDir(root), slug+".epic.json") }
func glensPath(root, slug string) string { return filepath.Join(backlogsDir(root), slug+".glens.json") }

// ---------------------------------------------------------------------------
// I/O
// ---------------------------------------------------------------------------

// validator is implemented by every entity that guards its own invariants.
type validator interface{ Validate() error }

// writeEntity validates (when supported) and writes v as tab-indented JSON,
// matching the api #015 example formatting.
func writeEntity(path string, v interface{}) error {
	if vv, ok := v.(validator); ok {
		if err := vv.Validate(); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// readEntity unmarshals the JSON document at path into v.
func readEntity(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// ---------------------------------------------------------------------------
// Venture resolution & collection loaders
// ---------------------------------------------------------------------------

// resolveVenture returns the venture root directory. An explicit path may point
// at the venture directory itself or at a parent holding exactly one; without a
// path, the current directory is probed the same way.
func resolveVenture(path string) (string, error) {
	if path != "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(venturePath(abs)); err == nil {
			return abs, nil
		}
		if v := findSingleVenture(abs); v != "" {
			return v, nil
		}
		return "", fmt.Errorf("no venture found at %s (missing venture.json)", abs)
	}
	cwd, _ := os.Getwd()
	if _, err := os.Stat(venturePath(cwd)); err == nil {
		return cwd, nil
	}
	if v := findSingleVenture(cwd); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("no .venture found in %s — run 'pft ideas new-venture <name>' or pass --path", cwd)
}

// findSingleVenture returns the sole *.venture child of dir, or "" if none or
// more than one exists.
func findSingleVenture(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var found []string
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".venture") {
			p := filepath.Join(dir, e.Name())
			if _, err := os.Stat(venturePath(p)); err == nil {
				found = append(found, p)
			}
		}
	}
	if len(found) == 1 {
		return found[0]
	}
	return ""
}

// globEntities returns sorted file paths in subdir matching suffix; a missing
// directory yields an empty slice, not an error.
func globEntities(root, subdir, suffix string) ([]string, error) {
	dir := filepath.Join(root, subdir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

func loadOpportunities(root string) ([]*Opportunity, error) {
	paths, err := globEntities(root, "records", ".opportunity.json")
	if err != nil {
		return nil, err
	}
	out := make([]*Opportunity, 0, len(paths))
	for _, p := range paths {
		o := &Opportunity{}
		if err := readEntity(p, o); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, o)
	}
	return out, nil
}

func loadUseCases(root string) ([]*UseCase, error) {
	paths, err := globEntities(root, "records", ".usecase.json")
	if err != nil {
		return nil, err
	}
	out := make([]*UseCase, 0, len(paths))
	for _, p := range paths {
		u := &UseCase{}
		if err := readEntity(p, u); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, u)
	}
	return out, nil
}

func loadEpics(root string) ([]*Epic, error) {
	paths, err := globEntities(root, "backlogs", ".epic.json")
	if err != nil {
		return nil, err
	}
	out := make([]*Epic, 0, len(paths))
	for _, p := range paths {
		e := &Epic{}
		if err := readEntity(p, e); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, e)
	}
	return out, nil
}

func loadInitiatives(root string) ([]*Initiative, error) {
	paths, err := globEntities(root, "initiatives", ".initiative.json")
	if err != nil {
		return nil, err
	}
	out := make([]*Initiative, 0, len(paths))
	for _, p := range paths {
		i := &Initiative{}
		if err := readEntity(p, i); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, i)
	}
	return out, nil
}

func loadBacklogs(root string) ([]*Backlog, error) {
	paths, err := globEntities(root, "backlogs", ".backlog.json")
	if err != nil {
		return nil, err
	}
	out := make([]*Backlog, 0, len(paths))
	for _, p := range paths {
		b := &Backlog{}
		if err := readEntity(p, b); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, b)
	}
	return out, nil
}

// findDiscoveryBacklog returns the sole discovery backlog in the venture. It is
// how add-idea locates where to register a new idea node.
func findDiscoveryBacklog(root string) (*Backlog, error) {
	backlogs, err := loadBacklogs(root)
	if err != nil {
		return nil, err
	}
	var found []*Backlog
	for _, b := range backlogs {
		if b.Kind == "discovery" {
			found = append(found, b)
		}
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("no discovery backlog in venture")
	case 1:
		return found[0], nil
	default:
		return nil, fmt.Errorf("multiple discovery backlogs; cannot pick one automatically")
	}
}

// ---------------------------------------------------------------------------
// Small utilities
// ---------------------------------------------------------------------------

// nowTS returns the current UTC time in the RFC3339 form used by the contract.
func nowTS() string { return time.Now().UTC().Format("2006-01-02T15:04:05Z") }

// strp returns a pointer to s (for nullable string fields set to a value).
func strp(s string) *string { return &s }

// f64p returns a pointer to f (for nullable number fields set to a value).
func f64p(f float64) *float64 { return &f }

// appendUnique appends val to slice unless already present (uniqueItems).
func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}

// splitCSV splits a comma-separated flag value into trimmed, non-empty tokens.
func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
