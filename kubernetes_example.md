Kubernetes Job Usage Example
============================

This example shows how **hvc** can be used as a Kubernetes Job. The example
shows both the **kubernetes** and **token** authentication strategies being 
employed in the **Copy Job Specification**.

The example assumes that the target Vault server has a Kubernetes Authentication
Method enabled at the standard path `auth/kubernetes` and it has a Role named
`hvc` defined that will allow the **hvc** Kubernetes Service Account to be used
to authenticate.

### Security Note

This example is tailored for simplicity over security, which is why a Kubernetes
Secret is created without any consideration to protect its value.  In a
production environment, the Kubernetes Secret should be created out-of-band and
proper measures should be taken to not disclose the secret's value.

```
$ echo 'apiVersion: v1
kind: ServiceAccount
metadata:
  name: hvc
---
apiVersion: v1
kind: Secret
metadata:
  name: source-vault-token
stringData:
  vault-token: s.0123456789abcdefghijklmn
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: spec
data:
  spec.json: |
    {
      "target": {
        "address": "http://target.vault:8200"
        "login": {
          "kubernetes": {
            "role": "hvc"
          }
        }
      },
      "sources": {
        "s1": {
          "address": "http://source.vault:8200"
          "login": {
            "token": "${SOURCE_VAULT_TOKEN}"
          }
        }
      },
      "copies": [
        {
          "path": "p1/secret"
          "secret": {
            "source": "s1",
            "path": "p1/secret"
          }
        }
      ]
    }
---
apiVersion: batch/v1
kind: Job
metadata:
  name: hvc
spec:
  completions: 1
  template:
    metadata:
    spec:
      volumes:
      - name: spec
        configMap:
          name: spec
          items:
          - key: spec.json
            path: ./spec.json
      containers:
      - name: hvc
        image: marcboudreau/hvc:latest
        args:
        - copy
        - /mnt/spec.json
        env:
        - name: SOURCE_VAULT_TOKEN
          valueFrom:
            secretKeyRef:
              name: source-vault-token
              key: vault-token
        volumeMounts:
        - name: spec
          mountPath: /mnt
      restartPolicy: Never
      serviceAccountName: hvc
' | kubectl apply -f -
```