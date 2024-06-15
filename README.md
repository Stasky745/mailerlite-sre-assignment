# mailerlite-sre-assignment
This is an assignment done for MailerLite's interview process.

## Description
This operator manages custom resources for configuring email sending and sending of emails via a transactional email provider like MailerSend. The operator works cross namespace, and sends from multiple providers such as MailerSend and Mailgun.

## Assignment Deliverables
### Container and Deployment manifests for the operator
The code for these is available in the [**Dockerfile**](./Dockerfile) and [**manifests/deployment.yaml**](./manifests/deployment.yaml).

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Quickstart
This shows a quickstart way of deploying and testing.

#### Setup
Clone this repository and install all the dependencies:

```bash
make
make manifests
make install
```

We can then create the image:

```bash
# To push to remote registry
# $ make docker-build docker-push IMG=example/my-operator:latest
# But I install locally 
$ make docker-build IMG=mailerlite.io/mail-operator:latest

# You can then deploy using the make command, or use the deployment manifest in the [`manifests`](./manifests/) folder.
# $ make deploy IMG=example/my-operator:latest
```

#### Create resources necessary

Create a Secret with the API token of the service you are gonna do:

```sh
kubectl create secret generic --from-literal=token=[API TOKEN]
```

Then create a EmailSenderConfig object. There's an example of the manifest in [`config/samples`](./config/samples/). 

```yaml
apiVersion: email.mailerlite.io/v1
kind: EmailSenderConfig
metadata:
  labels:
    app.kubernetes.io/name: mailerlite-sre-assignment
    app.kubernetes.io/managed-by: kustomize
  name: emailsenderconfig-sample
  namespace: default
spec:
  ApiTokenSecretRef: [Name of secret with API token]
  SenderEmail: sender@example.com
  Provider: mailersend # or mailgun
```

And deploy it.

You can then do the same with an Email object.

```yaml
apiVersion: email.mailerlite.io/v1
kind: Email
metadata:
  name: e-1
  namespace: default
spec:
  senderConfigRef: emailsenderconfig-sample # Name of the EmailSenderConfig object
  recipientEmail: recipient@example.com
  subject: This is an example subject
  body: This is an example body
```

You can then check the status of the email:

```
kubectl get mails e-1 -oyaml
```

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/mailerlite-sre-assignment:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/mailerlite-sre-assignment:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Assignment Structure
The API declarations are all in [`api/v1`](./api/v1/), and the code for the controllers are all in [`internal/controller`](./internal/controller/).

The extra code used to keep the structure tidy can be found in the [`pkg`](./pkg/) folder.

A deployment manifest can be found in the [`manifests`](./manifests/) folder.

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/mailerlite-sre-assignment:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/mailerlite-sre-assignment/<tag or branch>/dist/install.yaml
```

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

