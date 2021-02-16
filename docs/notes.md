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

### NEXT TODO:
Check out oc certificatesigningrequests command. Maybe we can create trusted certifcates that way?
