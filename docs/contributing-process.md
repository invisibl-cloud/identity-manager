## Project Management
The Code, our TODOs and Documentation is maintained on
[GitHub](https://github.com/invisible/identity-manager). All Issues
should be opened in that repository.

## Issues

Features, bugs and any issues regarding the documentation should be filed as
[GitHub Issue](https://github.com/invisible/identity-manager/issues) in
our repository. We use labels like `kind/feature`, `kind/bug`, `area/aws` to
organize the issues. Issues labeled `good first issue` and `help wanted` are
especially good for a first contribution. If you want to pick up an issue just
leave a comment.

## Creating a New Issue

If you've encountered an issue that is not already reported, please create an issue that contains the following:

- Clear description of the issue
- Steps to reproduce it
- Appropriate labels

## Building and testing locally

The project uses the `make` build system. It'll run code generators, tests and
static code analysis.

Building the operator binary and docker image:

```shell
make build
make docker-build IMG=identity-manager:latest
```

Run tests and lint the code:
```shell
make test
make lint
```

## Creating a Pull Request

Each new pull request should:

- Reference any related issues
- Add tests that show the issues have been solved
- Pass existing tests and linting
- Contain a clear indication of if they're ready for review or a work in progress
- Be up to date and/or rebased on the master branch
