The authorizations defined by the serviceaccount and roles in the security folder are required for deputy to access the deployments resource on the API server.

If you create these resources on Openshift, a token-secret is automatically generated. Querying and base64-decoding this will result an access token usable against the Openshift API.