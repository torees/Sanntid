How to use github basics:

cd <local folder>
git clone https://github.com/torees/Sanntid #get local copy from github
git init
git status #What are the changes?
#do programming and changes here
git status #see changes compared to master branch on github
git add . #add all changes to working tree
git remote add origin https://github.com/torees/Sanntid #maybe not necessary if already connected #with repo from <clone> command
git commit -m "which commit is this"
git push origin master #

to remove a file from local repo: delete manually or with git rm <name>
git add -u
#normal commit and push
done!


To work on a separate branch while doing radical changes or cooparating.
git checkout -b <newbranch> #creates a new branch with name newbranch
git checkout <whatbranch>  #switch between branches in local repository
#new branches need to be committed and pushed as normally to add to remote repo. 
git push origin <whatbranch>