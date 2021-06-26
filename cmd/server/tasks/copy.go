package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	k8s "github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/internal/trust"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"k8s.io/apimachinery/pkg/types"
)

type patch struct {
	Component string
	Platform  string
	Patch     k8s.DeploymentPatch
}

type platform struct {
	ServerUri string
	Namespace string
	Secret    string
	CAData    []byte
}

func (h *handler) doCopy(rw http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	app, _ := request.StringParam(params, "app")
	at, _ := request.TimeParam(params, "at")
	from, _ := request.StringParam(params, "from")
	to, _ := request.StringParam(params, "to")

	if at == nil {
		t := time.Now()
		at = &t
	}

	if err := h.copy(app, at, from, to); err != nil {
		httputil.WriteErrorResponse(rw, err)
	} else {
		rw.WriteHeader(http.StatusAccepted)
		rw.Write(nil)
	}
}

func (h *handler) copy(app string, at *time.Time, sourceEnv, targetEnv string) error {
	switch {
	case app == "":
		return httputil.BadRequest("Missing or invalid value for parameter 'app'")
	case sourceEnv == "":
		return httputil.BadRequest("Missing or invalid value for parameter 'from'")
	case targetEnv == "":
		return httputil.BadRequest("Missing or invalid value for parameter 'to'")
	}

	now := time.Now()

	source, err := h.getDeployments(app, sourceEnv, at)
	if err != nil {
		return err
	}

	target, err := h.getDeployments(app, targetEnv, &now)
	if err != nil {
		return err
	}

	for _, op := range createPatches(source, target) {
		p, err := h.getPlatform(targetEnv, op.Platform)
		if err != nil {
			return err
		}
		client, err := k8s.NewClient(p.ServerUri, p.Secret, p.CAData)
		if err != nil {
			return err
		}
		data, err := json.Marshal(op.Patch)
		if err != nil {
			return err
		}
		fmt.Println("TODO: wait for a deployment to complete before starting the next component update")
		fmt.Println("TODO: for that purpose you can monitor the pods resources associated with the deployment")
		fmt.Println("TODO: once the number of pods running with the new version is equal to the expected number we can continue")
		fmt.Println(targetEnv)
		fmt.Println(op.Component)
		fmt.Println(op.Platform)
		fmt.Println(string(data))
		d, err := client.AppsV1().Deployments(p.Namespace).Patch(op.Component, types.StrategicMergePatchType, data)
		if err != nil {
			return err
		}
		fmt.Println("Patch applied:", d.Spec.Template.Spec.Containers[0].Image)
	}

	return nil
}

func (h *handler) getPlatform(env, name string) (*platform, error) {
	row := h.db.QueryRow(`
        SELECT COALESCE(pf_api_server, ''), COALESCE(pf_namespace, ''), COALESCE(pf_secret, '')
          FROM platforms
         WHERE pf_env = $1 AND pf_name = $2
    `, env, name)
	p := platform{}
	if err := row.Scan(&p.ServerUri, &p.Namespace, &p.Secret); err != nil {
		return nil, err
	}

	fmt.Println("TODO: replace hard coded client certificate")
	p.CAData = trust.CAData

	return &p, nil
}

func (h *handler) getDeployments(app, env string, at *time.Time) (deployments, error) {
	rows, err := h.db.Query(`
      WITH slice AS (
        SELECT app_id _app_id, env_id _env_id, MAX(valid_from) _timestamp FROM apps_history
         WHERE app_id = $1 AND env_id = $2 AND valid_from <= $3::TIMESTAMP
         GROUP BY app_id, env_id
      )
      SELECT apps_history.component_id, components.name, platforms.name, apps_history.image_ref FROM slice
        JOIN apps_history ON apps_history.app_id = _app_id AND apps_history.env_id = _env_id AND apps_history.valid_from = _timestamp
        JOIN components ON components.id = apps_history.component_id
        JOIN platforms ON platforms.id = apps_history.platform_id
       WHERE image_ref IS NOT NULL
    `, app, env, at)

	if err != nil {
		return nil, err
	}

	result := make([]deployment, 0)
	for i := 0; rows.Next(); i++ {
		d := deployment{}
		if err := rows.Scan(&d.Id, &d.ComponentName, &d.PlatformName, &d.ImageRef); err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	return result, nil
}

func createPatches(source, target deployments) []patch {
	patches := []patch{}

	if source.Len() != target.Len() {
		fmt.Printf("source and target length do not match!!!")
		return patches
	}

	sort.Sort(source)
	sort.Sort(target)

	for i := 0; i < source.Len(); i++ {
		s, t := source[i], target[i]
		if s.ImageRef != t.ImageRef {
			if s.ComponentName != t.ComponentName {
				fmt.Println("source and target name do not match!")
				continue
			}
			fmt.Println("TODO: replace hard coded platform name")
			patches = append(patches, patch{Component: s.ComponentName, Platform: "minishift", Patch: k8s.CreateImagePatch(s.ComponentName, s.ImageRef)})
		}
	}

	return patches
}
