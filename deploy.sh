TOOLNAME="$1" or "campwiz-backend"
BRANCH="$2" or "main"
become "$TOOLNAME"
toolforge build start -L https://gitlab.wikimedia.org/nokibsarkar/campwiz-backend.git -b "$BRANCH"
toolforge jobs load jobs.yaml
toolforge jobs restart campwiz-task-manager
toolforge jobs restart campwiz-backend-readonly
toolforge jobs restart campwiz-backend-thing
exit
exit