become campwiz-backend
toolforge build start -L https://gitlab.wikimedia.org/nokibsarkar/campwiz-backend.git 
toolforge jobs load jobs.yaml
toolforge jobs restart campwiz-task-manager
toolforge jobs restart campwiz-backend-readonly
toolforge jobs restart campwiz-backend-thing
exit
exit