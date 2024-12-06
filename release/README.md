# Cutting New Releases

## Requirements
<!-- 
TODO: Make this more official and up to date
-->
- GitHub CLI (Insert Link Here)
  - User logged into CLI with write access to registry-support repo

## Process
<!-- 
TODO: Update this process for the various ways to run the script
-->
1. Determine version and type you wish to cut
   1. E.g. "I want to cut version 2.0.0, which is a Major release" 
   2. Or "I want to cut version 2.1.0, which is a Minor release"
2. Set the appropriate environment variables
   1. `VERSION`
   2. `RELEASE_TYPE`
   3. `RELEASE_CANDIDATE` - optional, defaults to `false` if unset
