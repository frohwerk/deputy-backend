package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"

	k8s "github.com/frohwerk/deputy-backend/internal/kubernetes"
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
	app := params.Get("app")
	from := params.Get("from")
	to := params.Get("to")

	if err := h.copy(app, from, to); err != nil {
		httputil.WriteErrorResponse(rw, err)
	} else {
		rw.WriteHeader(http.StatusAccepted)
		rw.Write(nil)
	}
}

func (h *handler) copy(app, sourceEnv, targetEnv string) error {
	switch {
	case app == "":
		return httputil.BadRequest("Missing or invalid value for parameter 'app'")
	case sourceEnv == "":
		return httputil.BadRequest("Missing or invalid value for parameter 'from'")
	case targetEnv == "":
		return httputil.BadRequest("Missing or invalid value for parameter 'to'")
	}

	source, err := h.getDeployments(app, sourceEnv)
	if err != nil {
		return err
	}

	target, err := h.getDeployments(app, targetEnv)
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
	cafile := "E:/projects/go/src/github.com/frohwerk/deputy-backend/certificates/minishift.crt"
	cadata, err := os.ReadFile(cafile)
	if err != nil {
		return nil, err
	}
	p.CAData = cadata

	return &p, nil
}

func (h *handler) getDeployments(app, env string) (deployments, error) {
	rows, err := h.db.Query(`
		SELECT c.name, COALESCE(d.image_ref, '')
		  FROM apps_components a
		 CROSS JOIN platforms p
		 INNER JOIN components c ON c.component_id = a.component_id
		  LEFT JOIN deployments d ON d.component_id = c.component_id AND d.platform_id = p.pf_id
		 WHERE a.app_id = $1 AND p.pf_env = $2
	`, app, env)

	if err != nil {
		return nil, err
	}

	result := make([]deployment, 0)
	for i := 0; rows.Next(); i++ {
		result = append(result, deployment{})
		if err := rows.Scan(&result[i].Name, &result[i].ImageRef); err != nil {
			return nil, err
		}
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
			if s.Name != t.Name {
				fmt.Println("source and target name do not match!")
				continue
			}
			fmt.Println("TODO: replace hard coded platform name")
			patches = append(patches, patch{Component: s.Name, Platform: "minishift", Patch: k8s.CreateImagePatch(s.Name, s.ImageRef)})
		}
	}

	return patches
}
