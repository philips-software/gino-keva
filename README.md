[![Build and Test](https://github.com/philips-software/gino-keva/actions/workflows/main.yml/badge.svg)](https://github.com/philips-software/gino-keva/actions/workflows/main.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/philips-software/gino-keva)](https://goreportcard.com/report/github.com/philips-software/gino-keva)

<!-- omit in toc -->

# Gino Keva - Git Notes Key Values

Gino Keva is a simple Key Value store built on top of Git Notes. The key/values are stored as json in a git note linked to the current commit. If there's not a note present yet, Gino Keva will walk back in the git history, and copy the first note found over to the current commit.

<!-- omit in toc -->

## Table of Contents

- [Gino Keva - Git Notes Key Values](#gino-keva---git-notes-key-values)
  - [Table of Contents](#table-of-contents)
  - [Use case](#use-case)
    - [Use case - Store new component version](#use-case---store-new-component-version)
    - [Use case - List all component versions corresponding for a certain commit](#use-case---list-all-component-versions-corresponding-for-a-certain-commit)
  - [Requirements](#requirements)
  - [How to use](#how-to-use)
    - [Warning: Push your changes](#warning-push-your-changes)
    - [Set key/value pairs](#set-keyvalue-pairs)
    - [Unset keys](#unset-keys)
    - [Use custom notes reference](#use-custom-notes-reference)
  - [FAQ](#faq)
    - [I need additional git configuration? How can I do that?](#i-need-additional-git-configuration-how-can-i-do-that)
    - [I need a custom output format](#i-need-a-custom-output-format)

## Use case

_Although Gino Keva was written with the below use case in mind, it intends to be a generic tool. Don't get discouraged if your intended use is very different. Instead feel free to open a ticket, so we can discuss if we can make it work._

The need for Gino Keva was born in an environment where ~20 components (some would call micro-services) live together in a single repository. Every component is deployed in a docker container; together they form an application/service. There's a single build pipeline that triggers upon any change. The pipeline will then fan out and trigger an independent build (and test) for each component impacted by the change. For each component, this results in a new docker container which is versioned and pushed to the registry. Once all components are rebuilt, the set of containers (of which some newly built) can be deployed and tested and eventually be promoted to production.

Due to the selective build mechanism, the versions of components are not coupled. Some will rarely change, others frequently. Now how to keep track of the set of containers that make up the application? It makes sense to keep this build metadata  inside the version control system, so we have it available for each commit that was built. But we'd hate to see the build pipeline polluting the git history with artificial commits. This is where Gino Keva was born.

<!-- omit in toc -->

### Use case - Store new component version

Gino Keva is used to store the newly built version of any component as a key/value pair in git notes, linked to commit it was built from: `COMPONENT_FOO=1.1.0`. If no prior notes were present, Gino Keva will search through history and find the nearest one first to be used as a starting point.

<!-- omit in toc -->

### Use case - List all component versions corresponding for a certain commit

For each deployment, the list of containers which make up the application is simply collected based on the output of `gino-keva list`:

| Before              | After                           |
| ------------------- | ------------------------------- |
| COMPONENT_FOO=1.0.0 | COMPONENT_FOO=1.1.0 (updated)   |
| COMPONENT_BAR=1.2.3 | COMPONENT_BAR=1.2.3 (untouched) |
| ....                | ....                            |

## Requirements

- Git CLI: Gino Keva uses the git CLI as installed on the host. Tested with version 2.32.0, however any recent version should do.

## How to use

See below examples on how to use gino-keva, or run `gino-keva --help` for help.

### Warning: Push your changes

By default, gino-keva will not push your changes to the upstream. You likely would like to change this behaviour by specifying `--push`, or setting the environment variable `GINO_KEVA_PUSH=1`.
If you do not do this, subsequent fetches will overwrite any local changes made.

### Set key/value pairs

The first time you set a key/value, gino-keva will warn you that no prior notes were found within the last 50 commits:

````console
foo@bar:~$ gino-keva set my-key my_value
WARN[0000] No prior notes found within maximum depth!    ref=gino_keva

You can continue to add other key/values:
```console
foo@bar (f10b970d):~$ gino-keva set counter 12
foo@bar (f10b970d):~$ gino-keva set foo bar
````

### List all key/value pairs

```console
foo@bar (f10b970d):~$ gino-keva list
COUNTER=12
FOO=bar
MY_KEY=my_value
```

gino-keva converts the key/value pairs into json and stores these in git notes, along with the souce commit hash. You can retrieve this format using the `--output=raw` flag:

```console
foo@bar (f10b970d):~$ gino-keva list --output=raw
{"COUNTER":{"data":"12","source":"f10b970d"},"FOO":{"data":"bar","source":"f10b970d"},"MY_KEY":{"data":"my_value","source":"f10b970d"}}

foo@bar (f10b970d):~$ git commit --allow-empty -m "Dummy commit"
foo@bar (a8517558):~$ gino-keva set pi 3.14
foo@bar (a8517558):~$ gino-keva list --output=json
{"COUNTER":{"data":"12","source":"f10b970d"},"FOO":{"data":"bar","source":"f10b970d"},"MY_KEY":{"data":"my_value","source":"f10b970d"},"PI":{"data":"3.14","source":"a8517558"}}
```

### Unset keys

Finally, you can unset keys using `unset`:

```console
foo@bar (a8517558):~$ gino-keva unset foo
foo@bar (a8517558):~$ gino-keva list
COUNTER=12
MY_KEY=my_value
PI=3.14
```

### Use custom notes reference

By default the notes are saved to `refs/notes/gino-keva`, but this can be changed with the `--ref` command-line switch. To store your key/value under `refs/notes/banana`:

```console
foo@bar (a8517558):~$ gino-keva --ref=banana set color yellow
```

## FAQ

### I need additional git configuration? How can I do that?

Since Gino Keva simply uses the git CLI, you can use (most of) the options it provides to set/override the configuration. You could either use `git config` to setup the system is desired, or use environment variables to achieve the same.

Example: Add a key/value pair as the "whatever \<whatever@example.com>" user

```
GIT_CONFIG_COUNT=2 \
GIT_CONFIG_KEY_0="user.name" GIT_CONFIG_VALUE_0="whatever" \
GIT_CONFIG_KEY_1="user.email" GIT_CONFIG_VALUE_1="whatever@example.com" \
gino-keva set foo bar
```

### I need a custom output format

Gino Keva supports just simple `key=value` format (default), or json (`--output=json`). However, you can parse the output in any format you'd like.

Example: Use gino-keva as part of a GitHub action:

```console
foo@bar:~$ gino-keva list | awk -F= '{print "::set-output name="$1"::"$2}'
::set-output name=COUNTER::12
::set-output name=MY_KEY::my_value
::set-output name=PI::3.14
```

Example: Use gino-keva as part of an Azure Devops pipeline:

```console
foo@bar:~$ gino-keva list | awk -F= '{print "##vso[task.setvariable variable="$1"]"$2}'
##vso[task.setvariable variable=COUNTER]12
##vso[task.setvariable variable=MY_KEY]my_value
##vso[task.setvariable variable=PI]3.14
```
