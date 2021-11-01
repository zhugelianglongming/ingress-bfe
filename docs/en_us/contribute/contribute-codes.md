# Contribute Code
This document explains how to contribute code

## Requirements of coding
- For code and comment, follow the [Golang style guide](https://github.com/golang/go/wiki/Style).
- Provide unit test for all code
- Pass all unit test
- Follow our [regulations of submmiting codes](https://www.bfe-networks.net/en_us/development/submit_pr_guide/)

## Code Develop
Below tutorial will guide you to submit code

1. [Fork](https://help.github.com/articles/fork-a-repo/)
   
    Go to [BFE Ingress Github][], click `Fork` button to create a repository in your own github space.
    
    >  `https://github.com/${USERNAME}/ingress-bfe`
    
1. Clone
   
    Clone the repository in your own space to your local :
    ```bash
    $ git clone https://github.com/${USERNAME}/ingress-bfe
    $ cd ingress-bfe
    ```
   
1. Create local branch
   
    We currently use [Git Branching Model][] to develop, test, release and maintenance, refer to [Release Regulation][]。
    * all development for feature and bug fix should be performed in a new branch
    * create new branch from branch `develop` in most case

    Use `git checkout -b` to create and switch to a new branch.
    ```bash
    $ git checkout -b my-cool-stuff
    ```
    
    > Before checkout, verify by `git status` command and keep current branch clean, otherwise untracked files will be brought to the new branch. 
    
1. Use pre-commit hook

    We use [pre-commit][] tool to manage Git pre-commit hook. 
    
    
    1. run following command：
        ```bash
        $ pip install pre-commit
        $ pre-commit install
        ```
    1. use `gofmt` to format code.
    
1. Coding

1. Build and test

    Compile source code, build BFE Ingress Controller docker and then test it. See more instruction in [Deploy Guide](../deployment.md)
    
1. Commit

    run `git commit` .

    Provides commit message for each commit, to let other people know what is changed in this commit.`git commit` .
    
    <b> <font color="red">Notice：commit message is also required to trigger CI unit test，format as below:</font> </b>
    
    ```bash
    # trigger CI unit test in develop branch
    $ git commit -m "test=develop"
    
    # trigger CI unit test in release/1.1 branch
    $ git commit -m "test=release/1.1"
    ```
    
1. Keep local repository up-to-date

    An experienced Git user always pulls from the official repo before pushing. They even pulls daily or hourly, so they notice conflicts with others work earlier, it makes it easier to resolve conflicts which is smaller.

   ```bash
   git remote add upstream https://github.com/bfenetworks/bfe
   git pull upstream develop
   ```

10. Push to remote repository

    Push local to your repository on GitHub `https://github.com/${USERNAME}/ingress-bfe`

```bash
# Example: push to remote repository `origin` branch `my-cool-stuff`
$ git push origin my-cool-stuff
```

> Refer to BFE [Local Develop Guide](https://www.bfe-networks.net/en_us/development/local_dev_guide/)

## Pull Request

1. Create an Issue and finish Pull Request
1. Pass unit test
1. Delete the branch used at your own repository
1. Delete the branch used at your local repository

> Refer to BFE [Submit PR Guide][submit PR guide]

[BFE Ingress Github]: https://github.com/bfenetworks/ingress-bfe
[Git Branching Model]: http://nvie.com/posts/a-successful-git-branching-model/
[Release Regulation]: https://github.com/bfenetworks/bfe/blob/develop/docs/en_us/development/release_regulation.md
[pre-commit]: http://pre-commit.com/
[git remote]: https://git-scm.com/docs/git-remote
[submit PR guide]: https://www.bfe-networks.net/en_us/development/submit_pr_guide/
