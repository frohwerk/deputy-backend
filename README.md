## cmd/watch

Command line utility that watches deployments on a kubernetes namespace and writes a yaml file for each event. The file names contain a leading counter - padded with zeroes to the length of five - and the name of the deployment object. An output directory must be specified when the program is started. For details see `watch --help`

## cmd/importer - usage

Requires a configuration file `$HOME/.deputy`. Contains configuration data for clusters, including: a logical name, the base uri of the API server, the serviceaccount name and API token, and the certificates of trusted authorities for TLS connection to the cluster

Example content:

```yaml
default: minishift
clusters:
  minishift:
    host: https://192.168.178.31:8443
    account: system:serviceaccount:myproject:someaccount
    token: API-Token (JWT) goes here
    cadata: |-
      -----BEGIN CERTIFICATE-----
      MIIC6jCCAdKgAwIBAgIBATANBgkqhkiG9w0BAQsFADAmMSQwIgYDVQQDDBtvcGVu
      c2hpZnQtc2lnbmVyQDE2MDY5MDM1MjcwHhcNMjAxMjAyMTAwNTI2WhcNMjUxMjAx
      MTAwNTI3WjAmMSQwIgYDVQQDDBtvcGVuc2hpZnQtc2lnbmVyQDE2MDY5MDM1Mjcw
      ggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCjcUZvkgn1z6XoKCyMaKGf
      /rE3dv3st4cJpSv0HSclxFuI8VhmTcIGQnTo/OkcV/Yhvb9y694Xb2cG+nAggWgY
      KaRfiMW2qCFhExa2Iap1SELHHqEN8dl2ch0tSjF7+zAYfmbyI+dOH8xqzTzwhX4n
      S0g1+pw6KivYA1NjdTVFdgpWLewRK/DKgbArKg7teoU+nyplqah8oYKAGE0cC75H
      oS02ZePrbwVLzWq0dh7c/Heim3fi7RhA+w2LlL/GepqgrNZu05gFOWO2yFZU66Fa
      n1zz7VNNdzsRRihdCDnLrCvDnbflnKstERj4Uy9FFLdytRAGJxsmGKyb81/TkF9D
      AgMBAAGjIzAhMA4GA1UdDwEB/wQEAwICpDAPBgNVHRMBAf8EBTADAQH/MA0GCSqG
      SIb3DQEBCwUAA4IBAQAl9FqD8Zm0AoOUA4p7HfRMMepPXhWdRLSsiVMSSfI6WTw6
      2/mj+dLC1WKw9LvJdkHi3znEYS75KvPOiMITbAd4+MFuihQnjxOzsdDlVfK7Vpfw
      foLJn+FqHR4LcPZR/YSklQ9lTPjq+oUrEOxSMabGMUfHc5Ta/zkVFbSD/Esnts/U
      wpsglq/7UOCq3tMzUnMDwoKM/IW5Mv5GQUo4NOnlM4KO4iSx0LfYv+1lAPlugeFq
      WuFXFF/3kuZDJDFDaSQ0PzLs7Ha/RW/Ax+dtDUH5A21DsNLUHmJPFwlDmat61P+h
      MzwsyNwurz+AsIWOsEgbCRaBuLvYTyesdviOfiTR
      -----END CERTIFICATE-----
```