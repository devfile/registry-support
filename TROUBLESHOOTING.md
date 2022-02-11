# Troubleshooting the Devfile Registry

## Collecting Logs

Logs can be collected for both the devfile index server, and the OCI registry server in the devfile registry, that should assist with debugging any potential issues.

To retrieve the logs from the devfile index server:

```bash
kubectl logs <devfile-registry-pod> -c devfile-registry-bootstrap
```

To retrieve the logs from the oci registry server:

```bash
kubectl logs <devfile-registry-pod> -c oci-registry
```

## Potential Issues

### Devfile Registry Does Not Start

If the devfile registry fails to start, check the logs of both the devfile registry and oci registry containers. Any errors in the logs likely indicate that the devfile registry did not start properly.

Potential causes of this include:

- Specifying the wrong image for the devfile index image when deploying

    - **Note:** The devfile index image produced by the devfile registry build tool must be used here if deploying your own devfile registry.
    
- No stacks in the devfile index image

    - The devfile registry will fail to start if there are no devfile stacks to serve

### Devfile Stacks Not Available

If a given stack(s) in your devfile registry does not show up when the devfile registry is deployed, check the following:

1) Verify that the stack exists under the `stacks/` folder in your devfile registry repository
2) Verify that the devfile registry logs show it pushing the stack to the OCI registry
3) Verify that the index.json created by the registry build contains the devfile stack
4) Verify tha you specified the proper devfile index image when deploying the registry.
