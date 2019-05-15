# 158222 Assignment 4

[Project Board](https://app.asana.com/0/1123132487808262/board)

## Workflow:
This workflow assumes that the master branch is always production ready, and development should be done on development branches.

Clone the repo (Do this once to get setup):
```bash
git clone https://github.com/bballenn/Assignment4.git
```

Include any local files in the current (or nested) directory, files outside will not be included in the repo.

To make your master branch up to date with what is on github
```bash
git pull origin master
```

To start making your changes:
1. Create a development branch
  ``` bash
  git checkout -b my_branch_name
  ```

  *Do your awesome work!*

2. Merge your changes with master
This assumes you are currently on your development branch. If you are not here is how to change to it. (there is no -b needed as you are not creating a new branch). use `git branch` to list out your branches in case you've forgotten the name.

```bash
git checkout my_branch_name
```

For the commit message just include the ticket heading as details related to what you'll be working on will be here.
Add WIP (work in progress) or Done to the end of the ticket heading. These messages are just for our reference.

If you're using this method then make sure you have recently updated your master branch, being behind on the master branch may cause conflicts.

  ```bash
  # commit your work on your development branch, use the . or write individual file names
  git add .
  git commit -m "Ticket Heading (WIP|Done)"

  # commit your work on the master branch and update github repo
  git checkout master
  git pull
  git merge my_branch_name
  git push
  ```

  3. Congrats you're using git!
  Just repeat step 2, making sure you are update to date with master to avoid conflicts when merging