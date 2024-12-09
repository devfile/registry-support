# Cutting New Releases

## Requirements

- SSH key setup with GitHub
  - See [GitHub documentation](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account) for more information
- Write access to the [devfile/registry-support](https://github.com/devfile/registry-support) repository

## Process
<!-- 
TODO: Update this process for the various ways to run the script
-->
1. Determine version and type you wish to cut
   1. E.g. "I want to cut version 2.0.0, which is a Major release" 
   2. Or "I want to cut version 2.1.0, which is a Minor release"
2. Set the appropriate environment variables
   1. `VERSION`
        - In the form [Major].[Minor].[Patch]
   2. `RELEASE_TYPE`
        - One of [major, minor, patch]
   3. `RELEASE_CANDIDATE`
        - Defaults to `false` if unset
        - Only applicable for `major` release types

## Examples

Major release v1.1.1
```
export VERISON=1.1.1
export RELEASE_TYPE=major
bash release.sh
```

Major release v2.0.0 that is a release candidate
```
export VERSION=2.0.0
export RELEASE_CANDIDATE=true
export RELEASE_TYPE=major
bash release.sh
```

Minor release v2.1.0
```
export VERSION=2.1.0
export RELEASE_TYPE=minor
bash release.sh
```

Patch release v2.1.1
```
export VERSION=2.1.1
export RELEASE_TYPE=patch
bash release.sh
```