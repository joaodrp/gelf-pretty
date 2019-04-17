# Contributing

First off, thank you for considering contributing to this project. By participating you agree to abide our [code of conduct](/CODE_OF_CONDUCT.md).

## Bugs and Questions

If you have noticed a bug or have a question please ensure it was not already reported/asked by someone else. You can do this by [searching all issues](https://github.com/joaodrp/gelf-pretty/issues). 

If you are unable to find an issue addressing the same problem, please [open a new one][new issue]. Be sure to include a clear
  description and an executable test case demonstrating the expected behavior that is not occurring.

## New Features

If you want to implement a new feature please first discuss it by raising an issue.


## Development Instructions

### Fork and Branch

Please start by [forking this repository](https://help.github.com/en/articles/fork-a-repo) and creating a branch for your changes. 

The name of your branch should follow this syntax:

```
[fix|feature]/<issue number>-<short-description>
```

- Use `fix` as prefix if your changes address a bug, or `feature` if you are adding a new feature;
- Make sure to include the number of the issue that was previously raised to report the issue or propose the new feature;
- Add a short but accurate description of your fix or feature. 

Examples:

- `fix/23-colored-output-on-windows`
- `feature/42-new-special-app-field`

### Setup

This application is written in [Go](https://golang.org/).

Prerequisites:

* `make`
* [Go 1.12+](http://golang.org/doc/install)


Install the build and lint dependencies:

``` sh
$ make setup
```

Run the test suite to make sure everything is fine:

``` sh
$ make test
```

### Develop

Make your changes and try to build from source as you go:

``` sh
$ make build
```

When you are satisfied with the changes run the test suite and linters:

``` sh
$ make ci
```

#### Commit

Commit messages should be descriptive and well formatted. Have look at *[How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/#seven-rules)*.

### Create Pull Request

Push the changes to your fork and [create a Pull Request](https://help.github.com/en/articles/creating-a-pull-request) against the
master branch of this repository.

## Credits

### Contributors

Thank you to all the people who have already contributed to the project!

