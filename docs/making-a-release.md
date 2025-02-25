# Making a Contour release

This page documents the process for releasing a new version of Contour.

The release types are as follows. All are tagged from the same release branch.

- Alpha releases.
- Beta releases.
- RC (Release Candidate) releases.
- Final releases.
- Patch releases.

## Branch for release

As contours master branch is under active development, all releases are made from a branch.
Create a release branch locally, if it has not been created for a beta release, like so

```sh
% git checkout -b release-0.15
```

If you are doing a patch release on an existing branch, skip this step and just checkout the branch instead.

This branch is used for all the release types, alpha through final.
Each on of the release types is just a different tag on the release branch./

## Alpha, beta, and release candidates

The steps for an non-production release are

- Checkout the release branch

```sh
% git checkout -b release-0.15
```

- Tag the head of release branch with the relevant release tag (in this case `alpha.1`), and push

```sh
% git tag -a v0.15.0-alpha.1 -m 'contour 0.15.0 alpha 1'
% git push --tags
```

Once the tag is present on the release branch, Github Actions will build the tag and push it to Docker Hub for you.
Then, you are done until we move to a final release.

## Final release

### Release tag from release branch

Tag the head of your release branch with the release tag, and push

```sh
% git tag -a v0.15.0 -m 'contour 0.15.0'
% git push --tags
```

## Patch release

### Cherry-pick required commits into the release branch

Get any required changes into the release branch by whatever means you choose.

### Release tag from reelase branch

Tag the head of your release branch with the release tag, and push

```sh
% git tag -a v0.15.1 -m 'contour 0.15.1'
% git push --tags
```

## Finishing up

If you've made a production release (that is, a final release or a patch release), you have a couple of things left to do.

### Updating quickstart URL

The quickstart url, https://projectcontour.io/quickstart/contour.yaml redirects to the current stable release.
This is controlled by a line in `site/_redirects`. If the definition of `:latest` has changed, update the quickstart redirector to match.

### Do the Github release and write release notes

Now you have a tag pushed to Github, go to the release tab on github, select the tag and write up your release notes. For patch releases, include the previous release notes below the new ones.

### Toot your horn

- Post a note to the #contour channel on k8s slack, also update the /topic with the current release number
- Post a note to the #project-contour channel on the vmware slack, also update the /topic with the current release number
- Send an email to the project-contour mailing list
