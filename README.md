# NetworkService controller

This is a multus aware service controller.

*This is a temporary home for it, until it has been moved to a permanent with a
proper license.*

## Build

This project uses the operator-sdk, which you must have installed
locally. Build instructions are contained [here](docs/build.md).

Since there are no publicly available images (yet), you must build and push
images before you can install.

## Installation

To install, after building and pushing, edit the `image:` in
[manifests/install/operator.yaml](manifests/install/operator.yaml), and then
install as follows.

    kubectl apply -f manifest/install

## Usage

Sample `NetService` objects are contained in the [manifests](manifests)
directory. Each such object has the following fields.

- The name of the `NetService` object is the name of the resulting service that
  will be created.

- The `netAttachDef` field is the name of the network attachment definition to
  select on.

- The `selector` is the selector, with the same format as a normal service
  selector.

Upon creating your `NetService` object, a service with the same name will be
created, pulling the IPs from the corresponding network, and will be kept
updated as changes are made to pods.

## Design

Suppose the `NetService` object is called `example`. The operator creates three
things as dependencies of the `NetService` object.

- It creates a headless service called `example-template` with a selector as
  specified. This automatically leads the endpoints controller to create and
  populate an endpoints object, `example-template`.

- It creates a headless service called `example` with no selector.

- It creates an endpoints object called `example`. This endpoints object is
  kept up to date with the IPs of the correct network from the pods which
  appear in `example-template`.

## Limitations

There are some limitations.

- This isn't production quality and has had limited testing.

- The build process (install operator-sdk, copy in some files) should be
  cleaned up.

- It only supports headless services. It could clearly be extended to other
  types, but since not all networks are routable from everywhere in general,
  adding cluster IP support requires some thought.

- Only IPv4 addresses are included, and only the first one for the network
  attachment definition (if any).

- The naming (with a hardcoded `-template`) doesn't feel very robust, and the
  `NetService` name for the custom resource is probably not right either.

- The annotation parsing code is frankly a hack, which should probably be
  replaced with use of the [standard
  client](https://github.com/k8snetworkplumbingwg/network-attachment-definition-client).

- There are a couple of other gotchas in the code tagged with `FIXME`.
