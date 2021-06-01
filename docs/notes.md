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

Simple use case
===============
Image:
- /app/app.js
- /app/lib/util.js
- /app/other.txt
- /var/opt/stuff.txt

Archive:
- app.js
- lib/util.js

Found at different location
===========================
Image:
- /app/app.js
- /lib/util.js
- /var/opt/stuff.txt

Archive:
- app.js
- lib/util.js

Whiteouts in Docker Images
==========================
- Remove all children with a specific prefix (example: file /etc/stuff/test.txt):
  empty file named '/etc/stuff/.wh.test' to remove a 
- Remove all children within a directory:
  empty file named '/path/.wh..wh..opq' (opaque whiteout)

Experiment #1
-------------
```Dockerfile
FROM alpine:13.3
ADD test.txt /etc/stuff/
```
Results in layer:
- /etc/stuff/.wh..wh..opq
- /etc/stuff/test.txt

Experiment #2
-------------
Note: Base image 'temp' is the result of Experiment #1
```Dockerfile
FROM temp
RUN rm -rf /etc/stuff
```
Results in layer:
- /etc/.wh.stuff

Docker images built by jib-maven-plugin
=======================================
Missing files (created by maven-jar-plugin):
- META-INF/MANIFEST.MF
- META-INF/maven/de.frohwerk/hello-world/pom.properties
- META-INF/maven/de.frohwerk/hello-world/pom.xml

Layer n:
- /app/classes

Layer n-1:
- /app/resources

Layer n-2:
- app\libs

------------------------------------------------------------------------------------------------

TODO:
- Add deployments table to store different image-refs for different platforms
- k8swatcher: store platform relationship in database
- Add certificate column to platform(?)
- Add environment dropdown to unassigned component selection (application-view)
- Add list of ignored suffixes for component names (e.g.: -dev|-test|-si|-pentest)
- Connect applications to environments
- List application components within a specific environment
- Compare application components between two environments
- Watch pods and track exact image hash for components

Relationships:
1 Stage <-> n Environments
1 Environment <-> n Components
1 Application <-> n Components
1 Application <-> n Stages

OR:
n Platforms (API-Server + Namespace) <-> 1 Environment (Logical Name)
=> New entity: platform with many-to-one relationship to environment


Add environment => Test
Add platform => https://192.168.178.31:8443 | myproject | 

TODO: Add to documentation
- PATCH method for deployment resource!
  https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/
  https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#patch-deployment-v1-apps
- Development environment: Allow image pull accross multiple namespaces
  oc policy add-role-to-user system:image-puller system:serviceaccount:demo-prod:default --namespace=myproject
- Algorithm for version comparison
- Using wildcard certificates for development:
  https://github.com/jsha/minica
