# Offene Punkte

1. Wie mit Images umgehen, die mehrere Container enthalten?

# Artifactory

https://www.jfrog.com/confluence/display/JFROG/Webhooks

Webhook Attribute:
- URL: obvious
- Event: Selection of events for the specific webhook
- Secret Token: Used for authentication against the webhook (protect the webhook endpoint against fabricated malicious event messages)
- Custom Headers: Additional headers to send with the event http request

# Cloud Foundry

Audit events help Cloud Foundry operators monitor actions taken against resources (such as apps) via user or system actions:
https://docs.cloudfoundry.org/running/managing-cf/audit-events.html

# Development environment

## Minishift

### Creating multiple minishift clusters

```bash
$ minishift profile set <name>
$ minishift start
```

### Reading local kubeconfig (created by kubectl)

This is supposed to help understand the variables required to authenticate at the API server.

```go
func main() {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error loading config: %s\n", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(config)
}
```

### Authentication towards the Kubernetes API server

Supported forms:
- mTLS using a X.509 client certificate
- username + password
- bearer tokens

On a developer machine the simplest approach is to use the bearer token stored in $HOME/.kube/config. This however requires an interactive login using a kubectl or oc (Openshift 
client).

The subject Common Name of a user certificate is used to identify the Kubernetes or Openshift user. The subject Organization(s) identify group memberships.

https://kubernetes.io/docs/reference/access-authn-authz/authentication/#x509-client-certs

### Creating a miniature certificate authority for development

For the creation and signing the tool [cfssl](https://github.com/cloudflare/cfssl) is used:

```bash
$ go get -u github.com/cloudflare/cfssl/cmd/cfssl
$ go get -u github.com/cloudflare/cfssl/cmd/cfssljson
```

Just follow this guide: https://kubernetes.io/docs/concepts/cluster-administration/certificates/

### Creating a service account:

```bash
$ oc create serviceaccount deputy
$ oc policy add-role-to-user -n myproject view system:serviceaccount:myproject:deputy
```

https://docs.openshift.com/container-platform/3.11/dev_guide/service_accounts.html
https://docs.okd.io/3.11/admin_guide/manage_users.html

### TODO:
Check out oc certificatesigningrequests command. Maybe we can create trusted certifcates that way?

### Importer draft

database:
- add resourceVersion column to components?

importer:
- 

```bash
$ curl http://ocrproxy-myproject.192.168.178.31.nip.io/v2/myproject/ocrproxy/tags/list
{"name":"myproject/ocrproxy","tags":["latest"]}
```

```bash
curl -GLs -H "Authorization: Bearer %DRT%" "https://registry-1.docker.io/v2/adoptopenjdk/openjdk11/tags/list"
```

Watch Pods...

After pod change:
1. GET pod owner (replicaset)
2. GET replicaset owner (deployment)

...OR...

deployment.spec.selector. 

### Events emitted when a new deployment is created:

Q: Is object.objectMetadata.uid stable related to a resource?
A: Yes

1. ADDED apps.Deployment
   - spec.containers[n].image contains image link
   - status.replicas|updatedreplicas|readyreplicas|availablereplicas|unavailablereplicas is zero
2. MODIFIED apps.Deployment
3. MODIFIED apps.Deployment
   - status.unavailablereplicas: 0 -> 1
4. ADDED core.Pod
   - object.objectmetadata.labels match the Deployment.Spec.PodTemplate
   - containers[n].image contains image link
   - status.phase=Pending
5. MODIFIED core.Pod
6. MODIFIED core.Pod
7. MODIFIED apps.Deployment
   - replicas|updatedreplicas: 0 -> 1
8. MODIFIED core.Pod
   - status.phase: Pending -> Running
9. MODIFIED apps.Deployment
   - readyreplicas: 0 -> 1
   - availablereplicas: 0 -> 1
   - unavailablereplicas: 1 -> 0

How to handle these events:
- New deployment created -> Create Component entry (image = null)
- Pod reaches Running phase -> Update Component set image = status.containerStatus[n].imageid

### Useful attributes of Deployment and Pod

- `[apps.Deployment] .status.*replicas` are counters representing the number of replicas in specific states
- `[apps.Deployment] .spec.containers[n].image` contains image link
- `[core.Pod] .status.containerStatuses[n].image` contains image link
- `[apps.Deployment] .spec.selector.matchLabels` contains matching labels
- `[core.Pod] .objectmeta.labels` contains matching labels

### Loading images from the registry:

Using:
https://github.com/distribution/distribution/tree/v2.7.1/registry/client
https://pkg.go.dev/github.com/distribution/distribution@v2.7.1+incompatible/registry/client

1. Fetch manifest
-> Problem: UnmarshalFunc nicht registriert in Standard-Konfiguration...
-> Nope: Just use the manifest types in your code, an init function configures the UnmarshalFunc when a type is used
2. Fetch and analyze layers
-> OPEN: How to handle whiteouts in layers (example: .wh..wh..opq)
-> See spec: https://github.com/opencontainers/image-spec/blob/master/layer.md#whiteouts

IDEA #1: Put source references as text files into the image itself
IDEA #2: Build smart analyzers (later?) for specific types of images

###

ifs:
app.js  1
lib.js  2

ofs:
/etc/app.js  1
/app/app.js  1
/app/lib.js  2

1. indexFile := app.js 1 (shortest path, sorting irrelevant)
2. dirs := ofs.find(app.js 1) => dirs = ['/etc', '/app']
3. for each dir := range dirs
   if dir.contains(ifs) return true

dir.contains: Paths, filenames and hashes must match

------------------------

for each archive:
  search archive by name, digest in images
  if found return image

  search largest file by relative path & digest in images
  if not found remove image from candidates list
  search second largest file by relative path & digest in images
  ...until all files have been found

  TODO: Handle identical files in different locations, see filesystem_test.go


path_suffix: the relative path of the file
