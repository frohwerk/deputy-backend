package copy

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/frohwerk/deputy-backend/cmd/server/apps"
	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
	"github.com/frohwerk/deputy-backend/cmd/workshop/job"
	"github.com/frohwerk/deputy-backend/cmd/workshop/rollout"
	"github.com/frohwerk/deputy-backend/internal/epoch"
	"github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/frohwerk/deputy-backend/internal/logger"
)

var (
	Log logger.Logger = logger.Default

	validId = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
)

func Job(db *sql.DB, apps *apps.Repository, platforms *kubernetes.ConfigRepository, out job.Output) job.Runner {
	return &copy{db, apps, platforms, out}
}

type copy struct {
	db        *sql.DB
	apps      *apps.Repository
	platforms *kubernetes.ConfigRepository
	out       job.Output
}

func (job *copy) Run(p job.Params) error {
	appId := p.Get("appId")
	source := p.Get("source")
	target := p.Get("target")

	beforeString := p.Get("before")
	before, err := parseTime(beforeString)
	if err != nil {
		return fmt.Errorf("invalid parameter value 'at': %s", err)
	}

	job.out.Write("Source time: %v", before)

	if !validId.MatchString(appId) {
		name, row := appId, job.db.QueryRow(`SELECT id FROM apps WHERE name = $1`, appId)
		if err := row.Scan(&appId); err != nil {
			return fmt.Errorf("No application with name '%s' found", name)
		}
	}

	if !validId.MatchString(source) {
		name, row := appId, job.db.QueryRow(`SELECT id FROM envs WHERE name = $1`, source)
		if err := row.Scan(&source); err != nil {
			return fmt.Errorf("No application with name '%s' found", name)
		}
	}

	if !validId.MatchString(target) {
		name, row := appId, job.db.QueryRow(`SELECT id FROM envs WHERE name = $1`, target)
		if err := row.Scan(&target); err != nil {
			return fmt.Errorf("No application with name '%s' found", name)
		}
	}

	job.out.Write("app: %s", appId)
	job.out.Write("target: %s", target)
	job.out.Write("source: %s", source)

	planner := rollout.Strategy(dependencies.Lookup{Store: dependencies.Cache(dependencies.DefaultDatabase(job.db))})

	targetEnv, err := job.platforms.Environment(target)
	if err != nil {
		return fmt.Errorf("error reading target environment configuration: %s", err)
	}

	targetApp, err := job.apps.CurrentView(appId, target)
	if err != nil {
		return fmt.Errorf("error reading target application: %s", err)
	}

	sourceApp, err := job.apps.History(appId, source, &before)
	if err != nil {
		return fmt.Errorf("error reading source application: %s", err)
	}

	patches, err := job.createPatches(sourceApp.Components, targetApp.Components)
	if err != nil {
		return fmt.Errorf("error creating patches for target: %s", err)
	}

	sort.Slice(patches, func(i, j int) bool { return patches[i].ComponentId > patches[j].ComponentId })
	job.out.Write("Patches before planing stage: %s", rollout.PatchList(patches))

	plan, err := planner.CreatePlan(patches)
	if err != nil {
		return fmt.Errorf("error creating patches for target: %s", err)
	}

	if len(plan) > -2 {
		job.out.Write("Rollout plan: %s", plan)
		return nil
	}

	job.out.Write("Patching environment %s", target)
	for _, patch := range plan {
		target, err := targetEnv.Platform(patch.PlatformName)
		if err != nil {
			return fmt.Errorf("error reading target platform: %v", err)
		}

		// Maybe use context for timeout?
		complete, err := target.Apply(&patch)
		if err != nil {
			return fmt.Errorf("error applying patch: %v", err)
		}

		// Wait for completion
		<-complete
	}

	// TODO ./cmd/server/apps/get.go: Reuse apps view with history
	// Input: app_id, at (timestamp), from_env, to_env
	job.out.Write("TODO: add timeout for the whole thing")
	// For each component (determine a reasonable order!):
	//
	return nil
}

func (job *copy) createPatches(source, target []apps.Component) ([]kubernetes.DeploymentPatch, error) {
	if len(source) != len(target) {
		return nil, fmt.Errorf("source and target must have the same components")
	}

	patches := []kubernetes.DeploymentPatch{}
	sort.Sort(byId(source))
	sort.Sort(byId(target))

	job.out.Write("Components (source):")
	for _, c := range source {
		job.out.Write("- %s => %s", c.Name, c.Image)
	}

	job.out.Write("Components (target):")
	for _, c := range target {
		job.out.Write("- %s => %s", c.Name, c.Image)
	}

	for i := 0; i < len(source); i++ {
		source, target := source[i], target[i]
		switch {
		case source.Id != target.Id:
			return nil, fmt.Errorf("source and target must have the same components")
		case source.Platform != target.Platform:
			return nil, fmt.Errorf("source and target must use the same platform (may use different environments)")
		case source.Image == target.Image:
			Log.Debug("%v == %v", source.Image, target.Image)
			continue
		case source.Image == "":
			return nil, fmt.Errorf("source has no image specified for component %s", source.Id)
		case target.Image == "":
			return nil, fmt.Errorf("target has no image specified for component %s", source.Id)
		}
		patch := kubernetes.DeploymentPatch{
			ComponentId:   source.Id,
			ComponentName: source.Name,
			PlatformName:  source.Platform,
			Spec: kubernetes.DeploymentPatchSpec{
				Template: kubernetes.PodTemplatePatch{
					Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: source.Name, Image: source.Image}}},
				},
			},
		}
		patches = append(patches, patch)
	}

	return patches, nil
}

func parseTime(t string) (time.Time, error) {
	switch t {
	case "now":
		return time.Now(), nil
	default:
		return epoch.ParseTime(t)
	}
}

type byId []apps.Component

func (s byId) Len() int {
	return len(s)
}

func (s byId) Less(i, j int) bool {
	return s[i].Id < s[j].Id
}

func (s byId) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
