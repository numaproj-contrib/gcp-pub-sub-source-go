git filter-branch --commit-filter '
    git commit-tree "$@" -S;
' HEAD
