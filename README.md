# Generation of SLSA3+ provenance for Node.js packages

This repository contains a reference implementation for generating non-forgeable [SLSA provenance](https://slsa.dev/) that meets the requirement for the [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels) for projects using the Go programming language.

***Note: This is a beta release and we are looking for your feedback. The official 1.0 release should come out in the next few weeks*** 

________
[Generation of provenance](#generation)
- [Example provenance](#example-provenance)
- [Configuration file](#configuration-file)
- [Workflow inputs](#workflow-inputs)
- [Workflow Example](#workflow-example)

[Verification of provenance](#verification-of-provenance)
- [Inputs](#inputs)
- [Command line examples](#command-line-examples)

[Technical design](#technial-design)
- [Blog posts](#blog-posts)
- [Specifications](#specifications)
________

## Generation
To generate provenance for an npm package, follow the steps below:

### Example provenance
An example of the provenance generated from this repo is below:
```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "subject": [
    {
      "name": "binary-linux-amd64",
      "digest": {
        "sha256": "0ae7e4fa71686538440012ee36a2634dbaa19df2dd16a466f52411fb348bbc4e"
      }
    }
  ],
  "predicate": {
    "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator-go/.github/workflows/builder.yml@main"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator-go@v1",
    "invocation": {
      "configSource": {
        "uri": "git+https://github.com/asraa/slsa-on-github-test@refs/heads/main.git",
        "digest": {
          "sha1": "11dba28bf106e98f9992daa56e3967be41a5f11d"
        },
        "entryPoint": "Test SLSA"
      },
      "parameters": {
        "version": 1,
        "event_name": "workflow_dispatch",
        "ref_type": "branch",
        "ref": "refs/heads/main",
        "base_ref": "",
        "head_ref": "",
        "actor": "asraa",
        "sha1": "11dba28bf106e98f9992daa56e3967be41a5f11d",
        "event_payload": ...
      },
      "environment": {
        "arch": "amd64",
        "github_event_name": "workflow_dispatch",
        "github_run_attempt": "1",
        "github_run_id": "1995071837",
        "github_run_number": "95",
        "os": "ubuntu"
      }
    },
    "buildConfig": {
      "version": 1,
      "steps": [
        {
          "command": [
            "/opt/hostedtoolcache/go/1.17.7/x64/bin/go",
            "build",
            "-mod=vendor",
            "-trimpath",
            "-tags=netgo",
            "-ldflags=-X main.gitVersion=v1.2.3 -X main.gitSomething=somthg",
            "-o",
            "binary-linux-amd64"
          ],
          "env": [
            "GOOS=linux",
            "GOARCH=amd64",
            "GO111MODULE=on",
            "CGO_ENABLED=0"
          ]
        }
      ]
    },
    "materials": [
      {
        "uri": "git+asraa/slsa-on-github-test.git",
        "digest": {
          "sha1": "11dba28bf106e98f9992daa56e3967be41a5f11d"
        }
      }
    ]
  }
}
```

### Configuration file

Define a configuration file called `.slsa-nodereleaser.yml` in the root of your project:

```yml
# TODO: what should this file look like?
version: 1
```

### Workflow inputs

The builder workflow [bcoe/slsa-github-generator-node/.github/workflows/builder.yml](.github/workflows/builder.yml) accepts the following inputs:

| Name | Required | Description |
| ------------ | -------- | ----------- |
| `env` | no | A list of environment variables, seperated by `,`: `VAR1: value, VAR2: value`. This is typically used to pass dynamically-generated values, such as `max_old_space_size`. Note that only environment variables with names starting with `NODE_` or `NODE` are accepted.|

### Workflow Example
Create a new workflow, say `.github/workflows/slsa-nodereleaser.yml`:

```yaml
name: SLSA node releaser
on:
  workflow_dispatch:
  push:
    tags:
      - "*" 

permissions: read-all
      
jobs:
  # Trusted builder.
  build:
    permissions:
      id-token: write
      contents: read
    uses: bcoe/slsa-github-generator-node/.github/workflows/builder.yml@main # TODO: use hash upon release.
  
  # Upload to GitHub release.
  upload:
    permissions:
      contents: write
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/download-artifact@fb598a63ae348fa914e94cd0ff38f362e927b741
        with:
          name: ${{ needs.build.outputs.node-package-name }}
      - uses: actions/download-artifact@fb598a63ae348fa914e94cd0ff38f362e927b741
        with:
          name: ${{ needs.build.outputs.node-package-name }}.intoto.jsonl
      - name: Release
        uses: softprops/action-gh-release@1e07f4398721186383de40550babbdf2b84acfc5
        if: startsWith(github.ref, 'refs/tags/')
        with:
          files: |
            ${{ needs.build.outputs.node-package-name }}
            ${{ needs.build.outputs.node-package-name }}.intoto.jsonl
      - uses: actions/checkout@a12a3943b4bdde767164f792f33f40b04645d846 # v2.3.4
        # these if statements ensure that a publication only occurs when
        # a new release is created:
        if: ${{ steps.release.outputs.release_created }}
      - name: Publish
        uses: actions/setup-node@56337c425554a6be30cdef71bf441f15be286854 # v3.1.1
        with:
          node-version: 16
          registry-url: 'https://wombat-dressing-room.appspot.com'
      - run: npm publish ${{ needs.build.outputs.node-package-name }}
        env:
          NODE_AUTH_TOKEN: ${{secrets.NPM_TOKEN}}
        if: ${{ steps.release.outputs.release_created }}
```

## Verification of provenance
To verify the provenance, use the [github.com/slsa-framework/slsa-verifier](https://github.com/slsa-framework/slsa-verifier) project. 

### Inputs
```shell
$ git clone git@github.com:slsa-framework/slsa-verifier.git
$ go run . --help
    -binary string
    	path to a binary to verify
    -branch string
    	expected branch the binary was compiled from (default "main")
    -provenance string
    	path to a provenance file
    -source string
    	expected source repository that should have produced the binary, e.g. github.com/some/repo
    -tag string
    	[optional] expected tag the binary was compiled from
    -versioned-tag string
    	[optional] expected version the binary was compiled from. Uses semantic version to match the tag
```

### Command line examples
```shell
$ go run . --binary ~/Downloads/binary-linux-amd64 --provenance ~/Downloads/binary-linux-amd64.intoto.jsonl --source github.com/origin/repo

Verified against tlog entry 1544571
verified SLSA provenance produced at 
 {
        "caller": "origin/repo",
        "commit": "0dfcd24824432c4ce587f79c918eef8fc2c44d7b",
        "job_workflow_ref": "/slsa-framework/slsa-github-generator-go/.github/workflows/builder.yml@refs/heads/main",
        "trigger": "workflow_dispatch",
        "issuer": "https://token.actions.githubusercontent.com"
}
successfully verified SLSA provenance
```

## Technical design

### Blog post
Find our blog post series [here](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

### Specifications
For a more in-depth technical dive, read the [SPECIFICATIONS.md](./SPECIFICATIONS.md).
