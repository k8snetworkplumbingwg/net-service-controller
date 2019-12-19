# Building this thing

## Creating the operator SDK structure

This follows the instructions [in the operator SDK
repo](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md).

- Make sure you have a recent version of go installed.

        plw@PC4003:~/work/go$ go version
        go version go1.13 linux/amd64

- Install the `operator-sdk` binary and put on your path.

- Ensure that `GOPATH` is set, and that `GO111MODULE` is set.

        export GO111MODULE=on

- Create the project.

        mkdir -p $GOPATH/src/operator-sdk
        cd $GOPATH/src/operator-sdk
        operator-sdk new svcctl
        cd svcctl

## Adding objects

- Add an API object for a `NetService`, and a controller for that object.

        operator-sdk add api --api-version=svcctl.metaswitch.com/v1alpha1 --kind=NetService
        operator-sdk add controller --api-version=svcctl.metaswitch.com/v1alpha1 --kind=NetService

- Copy in the two files from the [src](src} directory to replace these two files:

        $GOPATH/src/operator-sdk/svcctl/pkg/apis/svcctl/v1alpha1/netservice_types.go 
        $GOPATH/src/operator-sdk/svcctl/pkg/controller/netservice/netservice_controller.go

- Every time you change anything in the CRD definition, update the APIs etc.

        operator-sdk generate k8s && operator-sdk generate openapi

- To build the actual code (as opposed to CRDs) try this, where IMAGE should be
  a location you are allowed to push to:

        IMAGE=metaswitchglobal.azurecr.io/plw/svcctl-operator:latest
        operator-sdk build $IMAGE && docker push $IMAGE

- You need to set that image in the operator YAML too.

        sed -i "s|image:.*|image: ${IMAGE}|" manifests/install/operator.yaml

