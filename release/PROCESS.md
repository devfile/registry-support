# Release Process

## Versioning

The [`devfile/registry-support`](https://github.com/devfile/registry-support) repository has three kinds of releases and follows the [SemVer](https://semver.org/) format. Following this format we have the following:

- v[major].[minor].[patch]
- v[major].[minor].[patch]-rc for release-candidates

**Major** releases occur when there are changes towards the functionality of the registry support tool. Typically major releases will include significant feature changes with no guarantee of compatibility (usually part of a milestone), and changes from previous minor and patch releases. In addition to this, whenever a change is made to the API that breaks functionality, a major release will be cut. When a major release is cut there is no guarantee that prior Golang versions will or can be supported.

When a new release is cut the previous release will receive a dedicated release branch. For example, when `v2.1.0` is cut, the previous release, `v2.0.0` will receive a new branch named `release/v2.0.x`.

**Minor** releases occur when minor bug fixes, security patches, and regular feature changes are added. In addition, a minor release occurs when a new Golang version is released. Similiar to major releases, minor releases will receive a dedicated backport branch.

**Patch** releases only occur if a release needs to be cut outside of the normal release schedule/process. Patches should *only* include **critical** bug fixes and **critical** security patches that do not break the current release. Patches are tied to the latest minor release and are strongly recommended to end users. These patch releases have the potential to be backported to prior releases supporting different Golang versions.

**Pre-releases** happen for planned upcoming major releases to ensure all changes work as intended and gives a window for earlier adopters to try out the new major version. These pre-releases will receive a dedicated branch and will be post-fixed with `-rc`. For example, for a release `v3.0.0` that is marked as pre-release, a dedicated branch will be created named `rc/v3.0.0` and will be tagged `v3.0.0-rc`.

## Schedule

Major releases will be cut on an as-needed basis when major changes are made to how the application works.

Minor releases will roughly follow the release schedule of Golang, however, releases for feature additions, security fixes, and more can also be done on an as-needed basis.

## Changes

Release changes will consist of the merged PRs since the previous release that are made to the `main` branch. Patch changes made to a specific release branch would need to be backported to prior releases if necessary and the versioning would be adjusted accordingly. 

### Noteworthy Changes

Noteworthy changes should have the following criteria:
- Features should only include changes which directly impacts the user
- Go version should include any updates regarding a new go version being supported
- Bug fixes should include changes reported outside the team
- (Optional) Security Patches should include all changes
    - **Note**: Process of labelling security patches needs to be in place before we can identify them in release announcements, leaving as optional to the assignee’s discretion
- Documentation should include any change reported outside the team or highlights a feature of note

Changes within PRs can be highlighted as well with the PR as a base change.


## Cutting Releases

Individuals performing releases can find more information related to the process below.

### Major Releases

- See the dedicated release documentation [here](./README.md).

### Minor Releases

- See the dedicated release documentation [here](./README.md).

### Patch Releases

- See the dedicated release documentation [here](./README.md).
- If necessary, backport the change to the previous 2 releases.